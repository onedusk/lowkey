#!/usr/bin/env ruby
# frozen_string_literal: true

require 'thor'
require 'fileutils'

# File monitor CLI for tracking file changes in directories
#
# Monitors directories for file changes (new, modified, deleted) and logs
# all events with timestamps. Supports foreground and background monitoring.
#
# @example Watch directories in foreground
#   ./lowkey.rb watch ~/Documents ~/Downloads
#
# @example Start background monitoring
#   ./lowkey.rb start ~/Projects -i 10
#
# @example View logs filtered by event type
#   ./lowkey.rb log MODIFIED
#
# @example Follow logs in real-time
#   ./lowkey.rb tail
class Lowkey < Thor
  # Configuration constants
  MONITOR_DIR = File.join(Dir.home, '.file_monitor')
  PID_FILE = File.join(MONITOR_DIR, 'monitor.pid')
  CONFIG_FILE = File.join(MONITOR_DIR, 'config')
  DEFAULT_INTERVAL = 5

  # ANSI color codes
  COLORS = {
    red: "\e[0;31m",
    green: "\e[0;32m",
    yellow: "\e[0;33m",
    blue: "\e[0;34m",
    magenta: "\e[0;35m",
    reset: "\e[0m"
  }.freeze

  def initialize(*args)
    super
    FileUtils.mkdir_p(MONITOR_DIR)
    @quiet = false
    @interval = DEFAULT_INTERVAL
    @last_log_time = nil
  end

  # Configure Thor to exit with non-zero status on failures
  #
  # @return [Boolean] true to exit with non-zero status on errors
  def self.exit_on_failure?
    true
  end

  desc 'watch DIRS...', 'Start monitoring directories in foreground'
  method_option :interval, type: :numeric, aliases: '-i', desc: 'Check interval in seconds', default: DEFAULT_INTERVAL
  method_option :quiet, type: :boolean, aliases: '-q', desc: 'Quiet mode (no console output)', default: false

  # Watch directories for file changes in foreground
  #
  # @param dirs [Array<String>] directories to monitor
  # @option options [Integer] :interval check interval in seconds
  # @option options [Boolean] :quiet suppress console output
  # @return [void]
  def watch(*dirs)
    validate_directories!(dirs)

    @interval = options[:interval]
    @quiet = options[:quiet]

    save_config(dirs)
    setup_signal_handlers

    monitor_directories(dirs)
  end

  desc 'start DIRS...', 'Start monitoring directories in background'
  method_option :interval, type: :numeric, aliases: '-i', desc: 'Check interval in seconds', default: DEFAULT_INTERVAL

  # Start monitoring directories in background
  #
  # @param dirs [Array<String>] directories to monitor
  # @option options [Integer] :interval check interval in seconds
  # @return [void]
  def start(*dirs)
    if monitor_running?
      say "Monitor already running (PID: #{read_pid})", :red
      exit 1
    end

    validate_directories!(dirs)

    say 'Starting monitor in background...', :blue

    pid = Process.fork do
      Process.daemon(true)
      @interval = options[:interval]
      @quiet = true
      monitor_directories(dirs)
    end

    File.write(PID_FILE, pid)

    sleep 1

    if monitor_running?
      say "Monitor started (PID: #{pid})", :green
      say "Monitoring: #{dirs.join(', ')}"
      say "Use 'Lowkey tail' to follow logs"
    else
      say 'Failed to start monitor', :red
      FileUtils.rm_f(PID_FILE)
      exit 1
    end
  end

  desc 'stop', 'Stop background monitoring'

  # Stop background monitoring process
  #
  # @return [void]
  def stop
    unless monitor_running?
      say 'No monitor running', :yellow
      exit 1
    end

    pid = read_pid
    begin
      Process.kill('TERM', pid)
      FileUtils.rm_f(PID_FILE)
      say "Monitor stopped (PID: #{pid})", :green
    rescue Errno::ESRCH
      say 'Monitor not running (cleaning up stale PID file)', :yellow
      FileUtils.rm_f(PID_FILE)
    end
  end

  desc 'status', 'Show monitor status'

  # Display current monitor status
  #
  # @return [void]
  def status
    if monitor_running?
      pid = read_pid
      say "Monitor is running (PID: #{pid})", :green

      if File.exist?(CONFIG_FILE)
        say "\nMonitoring directories:"
        File.readlines(CONFIG_FILE).each { |dir| say "  • #{dir.strip}" }
      end

      log_files = Dir.glob(File.join(log_dir, '*.log')).sort
      if log_files.any?
        total = log_files.sum { |f| File.readlines(f).count { |line| !line.strip.empty? } }
        say "\nTotal events logged: #{total}"

        say "\nLast 5 events:"
        get_all_log_lines.last(5).each do |line|
          print_colored_log_line(line)
        end
      end
    else
      say 'Monitor is not running', :red
    end
  end

  desc 'log [PATTERN]', 'View logs with optional grep pattern'

  # Display logs, optionally filtered by pattern
  #
  # @param pattern [String, nil] optional pattern to filter logs
  # @return [void]
  def log(pattern = nil)
    logs = get_all_log_lines
    if logs.empty?
      say 'No logs found', :yellow
      exit 1
    end

    logs = logs.grep(/#{pattern}/i) if pattern
    logs.each { |line| print_colored_log_line(line) }
  end

  desc 'tail', 'Follow logs in real-time'

  # Follow logs in real-time (like tail -f)
  #
  # @return [void]
  def tail
    log_files = Dir.glob(File.join(log_dir, '*.log')).sort
    if log_files.empty?
      say 'No logs found', :yellow
      exit 1
    end

    say 'Following log files (Ctrl+C to stop)...', :blue

    # Start from the most recent log file
    current_file = log_files.last
    File.open(current_file) do |file|
      file.seek(0, IO::SEEK_END)

      loop do
        line = file.gets
        if line
          print_colored_log_line(line)
        else
          # Check if a new log file was created (new day)
          latest_files = Dir.glob(File.join(log_dir, '*.log')).sort
          if latest_files.last != current_file
            current_file = latest_files.last
            file.close
            file = File.open(current_file)
          end
          sleep 0.5
        end
      end
    end
  end

  desc 'summary', 'Show change statistics'

  # Display summary statistics of logged events
  #
  # @return [void]
  def summary
    logs = get_all_log_lines
    if logs.empty?
      say 'No logs found', :yellow
      exit 1
    end

    total = logs.size
    new_count = logs.count { |l| l.include?('[NEW]') }
    mod_count = logs.count { |l| l.include?('[MODIFIED]') }
    del_count = logs.count { |l| l.include?('[DELETED]') }

    say '=== File Monitor Summary ===', :blue
    say "Total events: #{total}", :magenta
    say "  New files:      #{new_count}", :green
    say "  Modified files: #{mod_count}", :yellow
    say "  Deleted files:  #{del_count}", :red

    show_most_active_files(logs)
    show_activity_by_hour(logs)
  end

  desc 'clear', 'Clear all logs'

  # Clear all log entries
  #
  # @return [void]
  def clear
    log_files = Dir.glob(File.join(log_dir, '*.log'))
    if log_files.any?
      log_files.each { |f| File.delete(f) }
      say "Logs cleared (#{log_files.size} files removed)", :green
    else
      say 'No logs to clear', :yellow
    end
  end

  no_commands do
    # Get the watched directory from config
    #
    # @return [String] first watched directory from config
    # @raise [Thor::Error] if no config exists
    def get_watch_dir
      if File.exist?(CONFIG_FILE)
        dirs = File.readlines(CONFIG_FILE).map(&:strip).reject(&:empty?)
        return File.expand_path(dirs.first) unless dirs.empty?
      end

      raise Thor::Error, 'No monitored directories configured'
    end

    # Get the log directory for the current watched directory
    #
    # @return [String] path to the log directory
    def log_dir
      watch_dir = get_watch_dir
      File.join(watch_dir, '.lowlog')
    end

    # Validate that all directories exist
    #
    # @param dirs [Array<String>] directories to validate
    # @raise [Thor::Error] if any directory doesn't exist
    # @return [void]
    def validate_directories!(dirs)
      if dirs.empty?
        say 'Error: No directories specified', :red
        exit 1
      end

      dirs.each do |dir|
        unless Dir.exist?(dir)
          say "Error: Directory '#{dir}' does not exist", :red
          exit 1
        end
      end
    end

    # Monitor directories for file changes
    #
    # @param dirs [Array<String>] directories to monitor
    # @return [void]
    def monitor_directories(dirs)
      states = {}

      abs_dirs = dirs.map { |d| File.expand_path(d) }

      say "Monitoring #{abs_dirs.size} directories", :blue unless @quiet
      say "Check interval: #{@interval}s", :blue unless @quiet
      say "Log directory: #{log_dir}", :blue unless @quiet
      say 'Press Ctrl+C to stop' unless @quiet
      say '---' unless @quiet

      abs_dirs.each do |dir|
        states[dir] = get_file_state(dir)
        say "Watching: #{dir}" unless @quiet
      end

      say '---' unless @quiet

      loop do
        sleep @interval

        abs_dirs.each do |dir|
          new_state = get_file_state(dir)

          if states[dir] != new_state
            compare_states(dir, states[dir], new_state)
            states[dir] = new_state
          end
        end
      end
    end

    # Get current state of all files in directory
    #
    # @param dir [String] directory path
    # @return [Hash] hash mapping file paths to [mtime, size]
    def get_file_state(dir)
      state = {}
      log_directory = File.join(dir, '.lowlog')

      Dir.glob(File.join(dir, '**', '*')).each do |path|
        next unless File.file?(path)
        next if path.start_with?(log_directory)

        begin
          stat = File.stat(path)
          state[path] = [stat.mtime.to_i, stat.size]
        rescue StandardError
          # Skip files we can't stat
        end
      end

      state
    end

    # Compare old and new states and log changes
    #
    # @param dir [String] directory being monitored
    # @param old_state [Hash] previous file state
    # @param new_state [Hash] current file state
    # @return [void]
    def compare_states(dir, old_state, new_state)
      # Find new and modified files
      new_state.each do |file, (mtime, size)|
        rel_path = file.sub("#{dir}/", '')

        if !old_state.key?(file)
          log_change('NEW', rel_path, "(#{size} bytes)")
        elsif old_state[file][0] != mtime
          old_size = old_state[file][1]
          size_diff = size - old_size
          size_info = size_diff.positive? ? "(+#{size_diff} bytes)" : "(#{size_diff} bytes)"
          log_change('MODIFIED', rel_path, size_info)
        end
      end

      # Find deleted files
      old_state.each_key do |file|
        unless new_state.key?(file)
          rel_path = file.sub("#{dir}/", '')
          log_change('DELETED', rel_path, '')
        end
      end
    end

    # Log a file change event
    #
    # @param event_type [String] type of event (NEW, MODIFIED, DELETED)
    # @param file [String] file path
    # @param details [String] additional details about the change
    # @return [void]
    def log_change(event_type, file, details)
      current_time = Time.now
      timestamp = current_time.strftime('%Y-%m-%d %H:%M:%S')
      log_entry = "[#{timestamp}] [#{event_type}] #{file} #{details}"

      # Get today's log file
      today = current_time.strftime('%Y-%m-%d')
      log_file = File.join(log_dir, "#{today}.log")

      # Ensure log directory exists
      FileUtils.mkdir_p(log_dir)

      # Check if we need to add gap (9 empty lines) for 1+ hour difference
      gap_needed = false
      gap_needed = true if @last_log_time && (current_time - @last_log_time) >= 3600 # 1 hour = 3600 seconds

      File.open(log_file, 'a') do |f|
        9.times { f.puts } if gap_needed
        f.puts(log_entry)
      end

      @last_log_time = current_time

      return if @quiet

      color = case event_type
              when 'NEW' then :green
              when 'MODIFIED' then :yellow
              when 'DELETED' then :red
              end

      say "[#{event_type[0..2]}] #{file} #{details}", color
    end

    # Print log line with appropriate color
    #
    # @param line [String] log line to print
    # @return [void]
    def print_colored_log_line(line)
      color = if line.include?('[NEW]')
                :green
              elsif line.include?('[MODIFIED]')
                :yellow
              elsif line.include?('[DELETED]')
                :red
              end

      if color
        say line.chomp, color
      else
        puts line
      end
    end

    # Check if monitor is currently running
    #
    # @return [Boolean] true if monitor is running
    def monitor_running?
      return false unless File.exist?(PID_FILE)

      pid = read_pid
      Process.kill(0, pid)
      true
    rescue Errno::ESRCH
      false
    end

    # Read PID from PID file
    #
    # @return [Integer] process ID
    def read_pid
      File.read(PID_FILE).strip.to_i
    end

    # Save configuration to file
    #
    # @param dirs [Array<String>] directories being monitored
    # @return [void]
    def save_config(dirs)
      File.write(CONFIG_FILE, dirs.join("\n"))
    end

    # Setup signal handlers for graceful shutdown
    #
    # @return [void]
    def setup_signal_handlers
      %w[INT TERM].each do |sig|
        Signal.trap(sig) do
          say "\nMonitoring stopped", :blue
          exit 0
        end
      end
    end

    # Display most active files from logs
    #
    # @param logs [Array<String>] log lines
    # @return [void]
    def show_most_active_files(logs)
      say "\nMost active files:", :blue

      file_counts = Hash.new(0)

      logs.each do |log|
        file_counts[Regexp.last_match(1)] += 1 if log =~ /\[(?:NEW|MODIFIED|DELETED)\]\s+(\S+)/
      end

      file_counts.sort_by { |_, count| -count }.first(5).each do |file, count|
        say "  #{count} changes: #{file}"
      end
    end

    # Display activity by hour from logs
    #
    # @param logs [Array<String>] log lines
    # @return [void]
    def show_activity_by_hour(logs)
      say "\nActivity by hour:", :blue

      hour_counts = Hash.new(0)

      logs.each do |log|
        hour_counts[Regexp.last_match(1)] += 1 if log =~ /^\[(\d{4}-\d{2}-\d{2} \d{2}):/
      end

      hour_counts.sort.last(5).each do |hour, count|
        say "  #{hour}:00  #{count} events"
      end
    end

    # Get all log lines from all log files, excluding empty lines
    #
    # @return [Array<String>] all log lines from all files
    def get_all_log_lines
      log_files = Dir.glob(File.join(log_dir, '*.log')).sort
      lines = []

      log_files.each do |file|
        File.readlines(file).each do |line|
          lines << line unless line.strip.empty?
        end
      end

      lines
    end
  end
end

Lowkey.start(ARGV) if __FILE__ == $PROGRAM_NAME
