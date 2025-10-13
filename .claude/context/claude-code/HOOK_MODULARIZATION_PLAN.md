# Claude Code Hook Modularization Plan

## Overview
This document outlines the plan to transform project-specific Claude Code hooks into a modular, reusable system that can be shared across all projects using the Claude Code SDK and uv package manager.

## Architecture

### Package Structure
```
claude-code-toolkit/              # Distributed Python package
├── src/
│   ├── hooks/                  # Hook implementations as Python modules
│   │   ├── validation.py       # Validation hook system
│   │   ├── git_sync.py         # Git synchronization
│   │   ├── context.py          # Context management
│   │   └── logging.py          # Command logging
│   ├── tools/                  # Custom MCP tools
│   │   ├── oober_tools.py      # Oober-specific tools
│   │   ├── validation_tools.py # Validation tools
│   │   └── dev_tools.py        # General dev tools
│   ├── agents/                 # Subagent definitions
│   │   ├── templates/          # Subagent markdown templates
│   │   └── installer.py        # Subagent installer
│   └── styles/                 # Output style definitions
│       ├── templates/          # Style markdown templates
│       └── installer.py        # Style installer
├── pyproject.toml              # Modern Python package config (uv compatible)
├── uv.lock                     # uv lock file
└── README.md
```

### Configuration Locations
- `~/.claude/` - User-level Claude Code configuration
- `.claude/` - Project-level configuration overrides
- `hooks.config.json` - Hook-specific configuration

## Key Components

### 1. SDK-Based Hook System
- Transform current file-based hooks into Python modules
- Use Claude Code SDK's `HookCallback` and `HookContext` interfaces
- Factory functions for easy hook registration
- Configuration hierarchy (defaults → user → project)

### 2. Custom MCP Tools
- Create reusable MCP tools using `@tool` decorator
- Package tools as SDK MCP servers
- Integrate with hooks for enhanced functionality

### 3. Subagent Templates
- Markdown templates for common subagent patterns
- Installer utility for deploying to `.claude/agents/`
- Support both user-level and project-level installation

### 4. Output Styles
- Predefined output style templates
- System prompt modifications for specialized behaviors
- Installer for `.claude/output-styles/`

### 5. Configuration System
- JSON-based configuration files
- Hierarchical override system
- Environment-specific settings

## Technology Stack

### Package Management: uv
- Fast, reliable Python package manager (10-100x faster than pip)
- Lock file support for reproducible installs
- Built-in virtual environment management
- Tool installation support for global commands
- Modern Python packaging standards (PEP 517/518/621)

### Dependencies
- `claude-code-sdk>=0.1.0` - Core SDK functionality
- `aiohttp>=3.8.0` - Async HTTP client
- `pyyaml>=6.0` - YAML configuration parsing

### Development Tools
- `pytest>=7.0` - Testing framework
- `pytest-asyncio>=0.21.0` - Async test support
- `black>=23.0` - Code formatting
- `ruff>=0.1.0` - Linting

## Installation Methods

### For End Users
```bash
# One-time global install
uv tool install claude-code-toolkit

# Use in any project
claude-toolkit init
```

### For Projects
```bash
# Add to project
uv add claude-code-toolkit

# Initialize
uv run claude-toolkit init --preset python
```

### For Development
```bash
# Clone and develop
git clone <repo>
cd claude-code-toolkit
uv sync
uv run pytest
```

## Migration Path

### Phase 1: Package Creation
- Set up project structure with uv
- Port existing hooks to Python modules
- Create MCP tool wrappers

### Phase 2: Configuration System
- Implement configuration loading hierarchy
- Create default configurations
- Add project type detection

### Phase 3: Installers
- Build subagent installer
- Build output style installer
- Create hook registration system

### Phase 4: Testing
- Unit tests for all components
- Integration tests with Claude Code SDK
- Cross-platform compatibility testing

### Phase 5: Distribution
- Publish to PyPI
- Create GitHub releases
- Write comprehensive documentation
- Provide installation guides

## Benefits

### Over Current Implementation
- **Reusability**: Single source of truth for all projects
- **Version Control**: Semantic versioning via package management
- **Testing**: Comprehensive test coverage
- **Type Safety**: Full Python type hints
- **Discovery**: Package registry and documentation

### uv Advantages
- **Speed**: 10-100x faster than pip
- **Reliability**: Better dependency resolution
- **Lock Files**: Reproducible installations
- **Tool Management**: Global tool installation support
- **Modern Standards**: Full PEP compliance

## Backwards Compatibility

The system maintains full compatibility with existing Claude Code patterns:
- Hooks still register in `settings.json`
- Subagents still use `.claude/agents/*.md`
- Output styles still use `.claude/output-styles/*.md`
- Configuration files remain JSON/YAML
- Can coexist with existing pip installations

## Usage Examples

### Basic Setup
```python
from claude_code_sdk import ClaudeSDKClient, ClaudeCodeOptions
from claude_code_toolkit import create_hooks, create_tools

# Auto-configure from project
hooks = create_hooks('.claude/hooks.config.json')
tools = create_tools(['validation', 'oober'])

options = ClaudeCodeOptions(
    hooks=hooks,
    mcp_servers={"toolkit": tools}
)
```

### Hook Configuration
```json
{
  "validation": {
    "enabled": true,
    "mode": "strict",
    "custom_rules": [
      {"pattern": "TODO", "message": "Use TODO: with colon"}
    ]
  },
  "git_sync": {
    "enabled": true,
    "exclude": ["*.log"]
  }
}
```

## Next Steps

1. Review and approve this plan
2. Create the `claude-code-toolkit` repository
3. Begin Phase 1 implementation
4. Set up CI/CD pipelines
5. Create initial documentation
6. Release alpha version for testing

---

*This plan leverages the Claude Code SDK's native capabilities while using uv's modern Python packaging for distribution and development.*