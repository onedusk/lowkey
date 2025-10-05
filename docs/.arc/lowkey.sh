#!/bin/bash

# File Monitor - All-in-one file change tracker for macOS
# Usage: ./monitor.sh [command] [directories...]

SCRIPT_NAME=$(basename "$0")
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_FILE="${HOME}/.file_monitor/changes.log"
PID_FILE="${HOME}/.file_monitor/monitor.pid"
STATE_FILE="${HOME}/.file_monitor/state"
CONFIG_FILE="${HOME}/.file_monitor/config"
CHECK_INTERVAL=5

# Ensure working directory exists
mkdir -p "${HOME}/.file_monitor"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Function to show usage
usage() {
    cat << EOF
File Monitor - Track file changes in directories

Usage: $SCRIPT_NAME [command] [options]

Commands:
    watch <dirs...>    Start monitoring directories
    start <dirs...>    Start monitoring in background
    stop              Stop background monitoring
    status            Show monitor status
    log [pattern]     View logs (optional grep pattern)
    tail              Follow logs in real-time
    summary           Show change statistics
    clear             Clear all logs
    
Options:
    -i <seconds>      Set check interval (default: 5)
    -q               Quiet mode (no console output)

Examples:
    $SCRIPT_NAME watch ~/Documents ~/Downloads
    $SCRIPT_NAME start ~/Projects -i 10
    $SCRIPT_NAME log MODIFIED
    $SCRIPT_NAME tail
    $SCRIPT_NAME stop

EOF
    exit 0
}

# Function to get file state (macOS compatible)
get_state() {
    local dir="$1"
    if [[ -d "$dir" ]]; then
        find "$dir" -type f -exec stat -f '%N|%m|%z' {} \; 2>/dev/null | sort
    fi
}

# Function to log changes
log_change() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local event_type="$1"
    local file="$2"
    local details="$3"
    local dir_name="$4"
    
    # Write to log file
    echo "[$timestamp] [$event_type] $file $details" >> "$LOG_FILE"
    
    # Console output (if not quiet)
    if [[ "$QUIET_MODE" != "true" ]]; then
        case "$event_type" in
            NEW)
                echo -e "${GREEN}[NEW]${NC} $file $details"
                ;;
            MODIFIED)
                echo -e "${YELLOW}[MOD]${NC} $file $details"
                ;;
            DELETED)
                echo -e "${RED}[DEL]${NC} $file"
                ;;
        esac
    fi
}

# Function to compare states
compare_states() {
    local dir="$1"
    local old_state="$2"
    local new_state="$3"
    local dir_name=$(basename "$dir")
    
    # Find new and modified files
    while IFS='|' read -r file mtime size; do
        [[ -z "$file" ]] && continue
        
        old_entry=$(echo "$old_state" | grep "^${file}|" || true)
        rel_path="${file#$dir/}"
        
        if [[ -z "$old_entry" ]]; then
            log_change "NEW" "$rel_path" "(${size} bytes)" "$dir_name"
        else
            old_mtime=$(echo "$old_entry" | cut -d'|' -f2)
            old_size=$(echo "$old_entry" | cut -d'|' -f3)
            
            if [[ "$mtime" != "$old_mtime" ]]; then
                if [[ "$size" != "$old_size" ]]; then
                    size_diff=$((size - old_size))
                    if [[ $size_diff -gt 0 ]]; then
                        size_info="(+${size_diff} bytes)"
                    else
                        size_info="(${size_diff} bytes)"
                    fi
                    log_change "MODIFIED" "$rel_path" "$size_info" "$dir_name"
                fi
            fi
        fi
    done <<< "$new_state"
    
    # Find deleted files
    while IFS='|' read -r file mtime size; do
        [[ -z "$file" ]] && continue
        
        if ! echo "$new_state" | grep -q "^${file}|"; then
            rel_path="${file#$dir/}"
            log_change "DELETED" "$rel_path" "" "$dir_name"
        fi
    done <<< "$old_state"
}

# Function to monitor directories
monitor_dirs() {
    local dirs=("$@")
    declare -A states
    
    # Validate directories
    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            echo "Error: Directory '$dir' does not exist"
            exit 1
        fi
    done
    
    # Save configuration
    printf "%s\n" "${dirs[@]}" > "$CONFIG_FILE"
    
    # Initial state
    echo -e "${BLUE}Monitoring ${#dirs[@]} directories${NC}"
    echo -e "${BLUE}Check interval: ${CHECK_INTERVAL}s${NC}"
    echo -e "${BLUE}Log file: $LOG_FILE${NC}"
    echo "Press Ctrl+C to stop"
    echo "---"
    
    # Get initial states
    for dir in "${dirs[@]}"; do
        abs_dir=$(cd "$dir" && pwd)
        states["$abs_dir"]=$(get_state "$abs_dir")
        echo "Watching: $abs_dir"
    done
    echo "---"
    
    # Monitoring loop
    while true; do
        sleep "$CHECK_INTERVAL"
        
        for dir in "${dirs[@]}"; do
            abs_dir=$(cd "$dir" && pwd)
            new_state=$(get_state "$abs_dir")
            
            if [[ "${states[$abs_dir]}" != "$new_state" ]]; then
                compare_states "$abs_dir" "${states[$abs_dir]}" "$new_state"
                states["$abs_dir"]="$new_state"
            fi
        done
    done
}

# Function to start monitoring in background
start_background() {
    if [[ -f "$PID_FILE" ]] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        echo "Monitor already running (PID: $(cat $PID_FILE))"
        exit 1
    fi
    
    echo "Starting monitor in background..."
    QUIET_MODE=true nohup "$0" watch "$@" > /dev/null 2>&1 &
    echo $! > "$PID_FILE"
    
    sleep 1
    if kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        echo "Monitor started (PID: $(cat $PID_FILE))"
        echo "Monitoring: $*"
        echo "Use '$SCRIPT_NAME tail' to follow logs"
    else
        echo "Failed to start monitor"
        rm -f "$PID_FILE"
        exit 1
    fi
}

# Function to stop background monitoring
stop_monitor() {
    if [[ ! -f "$PID_FILE" ]]; then
        echo "No monitor running"
        exit 1
    fi
    
    local pid=$(cat "$PID_FILE")
    if kill -0 "$pid" 2>/dev/null; then
        kill "$pid"
        rm -f "$PID_FILE"
        echo "Monitor stopped (PID: $pid)"
    else
        echo "Monitor not running (cleaning up stale PID file)"
        rm -f "$PID_FILE"
    fi
}

# Function to show status
show_status() {
    if [[ -f "$PID_FILE" ]] && kill -0 $(cat "$PID_FILE") 2>/dev/null; then
        local pid=$(cat "$PID_FILE")
        echo -e "${GREEN}Monitor is running${NC} (PID: $pid)"
        
        if [[ -f "$CONFIG_FILE" ]]; then
            echo -e "\nMonitoring directories:"
            while IFS= read -r dir; do
                echo "  • $dir"
            done < "$CONFIG_FILE"
        fi
        
        if [[ -f "$LOG_FILE" ]]; then
            local total=$(wc -l < "$LOG_FILE" | tr -d ' ')
            echo -e "\nTotal events logged: $total"
            
            echo -e "\nLast 5 events:"
            tail -5 "$LOG_FILE" | while IFS= read -r line; do
                if [[ "$line" == *"[NEW]"* ]]; then
                    echo -e "  ${GREEN}$line${NC}"
                elif [[ "$line" == *"[MODIFIED]"* ]]; then
                    echo -e "  ${YELLOW}$line${NC}"
                elif [[ "$line" == *"[DELETED]"* ]]; then
                    echo -e "  ${RED}$line${NC}"
                else
                    echo "  $line"
                fi
            done
        fi
    else
        echo -e "${RED}Monitor is not running${NC}"
    fi
}

# Function to show logs
show_logs() {
    local pattern="$1"
    
    if [[ ! -f "$LOG_FILE" ]]; then
        echo "No logs found"
        exit 1
    fi
    
    if [[ -n "$pattern" ]]; then
        grep -i "$pattern" "$LOG_FILE" | while IFS= read -r line; do
            if [[ "$line" == *"[NEW]"* ]]; then
                echo -e "${GREEN}$line${NC}"
            elif [[ "$line" == *"[MODIFIED]"* ]]; then
                echo -e "${YELLOW}$line${NC}"
            elif [[ "$line" == *"[DELETED]"* ]]; then
                echo -e "${RED}$line${NC}"
            else
                echo "$line"
            fi
        done
    else
        cat "$LOG_FILE"
    fi
}

# Function to tail logs
tail_logs() {
    if [[ ! -f "$LOG_FILE" ]]; then
        echo "No logs found"
        exit 1
    fi
    
    echo "Following log file (Ctrl+C to stop)..."
    tail -f "$LOG_FILE" | while IFS= read -r line; do
        if [[ "$line" == *"[NEW]"* ]]; then
            echo -e "${GREEN}$line${NC}"
        elif [[ "$line" == *"[MODIFIED]"* ]]; then
            echo -e "${YELLOW}$line${NC}"
        elif [[ "$line" == *"[DELETED]"* ]]; then
            echo -e "${RED}$line${NC}"
        else
            echo "$line"
        fi
    done
}

# Function to show summary
show_summary() {
    if [[ ! -f "$LOG_FILE" ]]; then
        echo "No logs found"
        exit 1
    fi
    
    local total=$(wc -l < "$LOG_FILE" | tr -d ' ')
    local new_count=$(grep -c "\[NEW\]" "$LOG_FILE" || echo 0)
    local mod_count=$(grep -c "\[MODIFIED\]" "$LOG_FILE" || echo 0)
    local del_count=$(grep -c "\[DELETED\]" "$LOG_FILE" || echo 0)
    
    echo -e "${BLUE}=== File Monitor Summary ===${NC}"
    echo -e "Total events: ${MAGENTA}$total${NC}"
    echo -e "  ${GREEN}New files:${NC}      $new_count"
    echo -e "  ${YELLOW}Modified files:${NC} $mod_count"
    echo -e "  ${RED}Deleted files:${NC}  $del_count"
    
    if [[ -f "$LOG_FILE" ]]; then
        echo -e "\n${BLUE}Most active files:${NC}"
        awk '{for(i=3;i<=NF;i++) if($i!~/^\(/) print $i}' "$LOG_FILE" | \
            sort | uniq -c | sort -rn | head -5 | \
            while read count file; do
                echo "  $count changes: $file"
            done
        
        echo -e "\n${BLUE}Activity by hour:${NC}"
        awk '{print substr($2,1,14)":00"}' "$LOG_FILE" | \
            sort | uniq -c | tail -5 | \
            while read count hour; do
                printf "  %s %4d events\n" "$hour" "$count"
            done
    fi
}

# Function to clear logs
clear_logs() {
    if [[ -f "$LOG_FILE" ]]; then
        > "$LOG_FILE"
        echo "Logs cleared"
    else
        echo "No logs to clear"
    fi
}

# Parse command line arguments
COMMAND="$1"
shift || true

# Parse options
QUIET_MODE=false
while [[ "$1" =~ ^- ]]; do
    case "$1" in
        -i)
            CHECK_INTERVAL="$2"
            shift 2
            ;;
        -q)
            QUIET_MODE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Execute command
case "$COMMAND" in
    watch)
        trap "echo -e '\n${BLUE}Monitoring stopped${NC}'; exit 0" INT TERM
        monitor_dirs "$@"
        ;;
    start)
        start_background "$@"
        ;;
    stop)
        stop_monitor
        ;;
    status)
        show_status
        ;;
    log|logs)
        show_logs "$1"
        ;;
    tail)
        tail_logs
        ;;
    summary)
        show_summary
        ;;
    clear)
        clear_logs
        ;;
    *)
        usage
        ;;
esac