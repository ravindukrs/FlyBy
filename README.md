# FlyBy - Concourse CI Terminal UI

FlyBy is a powerful Terminal User Interface (TUI) application that provides an intuitive and interactive interface for managing Concourse CI pipelines, jobs, builds, and resources. Built with Go and the Bubble Tea framework, it streamlines your Concourse workflow by replacing complex fly CLI commands with an elegant, keyboard-driven interface.

## ✨ Features

### 🎯 **Target Management**
- View and manage multiple Concourse targets
- Add new targets with interactive forms
- Automatic authentication handling
- Quick target switching

### 🚀 **Pipeline Operations**
- Browse all pipelines across teams
- View pipeline status (paused/unpaused)
- Trigger pipeline jobs
- Navigate to jobs and resources

### ⚙️ **Job Management**
- View all jobs within a pipeline
- Real-time job status monitoring
- One-click job triggering with live feedback
- Navigate to build history

### 🔨 **Build Operations** (NEW!)
- View complete build history for any job
- **Build Rerunning**: Re-run specific builds with the same inputs (just like Concourse web UI)
- Real-time build status and timing information
- Detailed build information display
- Auto-refresh after build operations

### 📊 **Resource Management**
- View all pipeline resources
- Check resources on-demand
- Real-time resource check feedback
- Last checked timestamps

### 🔐 **Authentication**
- Seamless authentication flow
- Automatic token management
- Handle expired sessions gracefully

### ⚡ **Real-time Operations**
- Live feedback for all operations
- Real-time status updates
- Loading indicators and progress feedback
- **Manual Refresh** with F5 key across all views

### 🔍 **Search Functionality** (NEW!)
- **Universal search** across all views (Targets, Pipelines, Jobs, Resources)
- **Real-time filtering** as you type
- **Multiple search fields**: Search by names, types, teams, and more
- **Visual search indicators** with active/inactive states
- **Keyboard shortcuts** for efficient search workflow

### 🎨 **User Experience**
- Intuitive keyboard navigation
- Color-coded status indicators
- Hierarchical navigation (Targets → Pipelines → Jobs/Resources → Builds)
- Comprehensive help text in footer

## 🚀 Installation

### From Source
```bash
git clone <repository-url>
cd FlyBy
go build -o build/flyby ./cmd/flyby
./build/flyby
```

### Using Make
```bash
make build
./build/flyby
```

## 📖 Usage

### Starting FlyBy
```bash
./build/flyby
```

### Navigation Structure
```
Main Menu
├── Targets (fly targets management)
│   ├── Select target → Pipelines
│   └── Add new target
│
├── Pipelines (for selected target)
│   ├── Jobs → Job list → Build history
│   └── Resources → Resource management
│
├── Jobs (for selected pipeline)
│   ├── Trigger job
│   └── View builds → Build rerunning
│
├── Resources (for selected pipeline)
│   └── Check resources
│
└── Builds (for selected job)
    └── Rerun specific builds
```

## ⌨️ Keyboard Controls

### Global Controls
- **Arrow Keys / j/k**: Navigate up/down
- **Enter**: Select/Confirm action
- **Esc**: Go back to previous view
- **q**: Quit application
- **F5**: Refresh current view ✨
- **/ or s**: Start search in any view ✨
- **Ctrl+U**: Clear search query ✨

### Search Mode Controls ✨
- **Type**: Enter search query (real-time filtering)
- **Enter**: Finish search (stay in filtered results)
- **Esc**: Cancel search and clear filter
- **Backspace**: Delete last character
- **Ctrl+U**: Clear entire search query

### Target View
- **a**: Add new target
- **d**: Delete target
- **Enter**: Select target and view pipelines
- **i**: Toggle detailed target information
- **/ or s**: Search targets by name, URL, or team

### Pipeline View
- **j**: View jobs for selected pipeline
- **r**: View resources for selected pipeline
- **p**: Pause/unpause pipeline
- **/ or s**: Search pipelines by name or team

### Jobs View
- **Enter/t**: Trigger selected job
- **b**: View build history for selected job
- **/ or s**: Search jobs by name, pipeline, or team

### Resources View
- **Enter/c**: Check selected resource
- **/ or s**: Search resources by name, type, pipeline, or team

### Builds View ✨
- **Enter**: **Rerun selected build** (with same inputs)
- **F5**: Refresh build list

## 🎯 Key Features Explained

### Universal Search System ✨

**FlyBy provides powerful search capabilities across all views:**

**Search Triggers:**
- Press **`/`** or **`s`** to start searching in any view
- Search box appears with real-time visual feedback

**Search Capabilities:**
- **Targets**: Search by name, API URL, or team
- **Pipelines**: Search by pipeline name or team  
- **Jobs**: Search by job name, pipeline, or team
- **Resources**: Search by resource name, type, pipeline, or team

**Search Features:**
- **Real-time filtering**: Results update as you type
- **Visual indicators**: Search box highlights when active
- **Cursor display**: Shows typing position in search mode
- **Selection preservation**: Maintains correct selection after filtering

**Search Workflow:**
1. **Start**: Press `/` or `s` to enter search mode
2. **Type**: Enter your search query (case-insensitive)
3. **Filter**: Results update in real-time
4. **Navigate**: Use arrow keys to select from filtered results
5. **Finish**: Press Enter to stay with filtered results
6. **Clear**: Press Esc to cancel search, or Ctrl+U to clear query

### Build Rerunning vs Job Triggering

**FlyBy distinguishes between two different operations:**

1. **Job Triggering** (Jobs view): Creates a new build with latest inputs
   - Uses `fly trigger-job` command
   - Gets latest resource versions
   - Creates new build number

2. **Build Rerunning** (Builds view): Reruns a specific build with identical inputs
   - Uses `fly rerun-build` command
   - Uses exact same resource versions as original build
   - Creates sub-build (e.g., #11.1, #11.2)
   - **Matches Concourse Web UI behavior**

### Real-time Feedback System

All operations provide comprehensive feedback:
- ✅ Success messages with command output
- ❌ Error messages with detailed information
- 🔄 Loading indicators during operations
- ⏱️ Automatic message cleanup after 5 seconds

### Refresh Functionality

Press **F5** in any view to refresh data:
- **Pipelines**: Reload pipeline list and status
- **Jobs**: Refresh job list and status  
- **Resources**: Update resource status and check times
- **Builds**: Reload build history (useful after build operations)

**Note**: Search filters are preserved during refresh operations.

### Search Examples

**Finding a specific pipeline:**
```
1. Navigate to Pipelines view
2. Press '/' to start search
3. Type "deploy" → Only pipelines with "deploy" in name/team show
4. Use arrows to select, Enter to proceed to jobs
```

**Filtering resources by type:**
```  
1. Navigate to Resources view
2. Press 's' to start search
3. Type "git" → Only git-type resources show
4. Select resource and press Enter to check
```

## 🔧 Requirements

- **Go 1.19+**: For building from source
- **fly CLI**: Must be installed and in PATH
- **Concourse Access**: Valid Concourse CI instance(s)
- **Terminal**: Modern terminal with color support

## ⚙️ Configuration

FlyBy reads your existing fly configuration:
- **Targets**: From `~/.flyrc`
- **Authentication**: Uses existing fly tokens
- **No additional setup required**

## 🏗️ Development

### Project Structure
```
FlyBy/
├── cmd/flyby/           # Main application entry point
├── internal/
│   ├── concourse/       # Concourse client and API wrapper
│   ├── config/          # Configuration management
│   └── tui/            # Terminal UI components
├── build/              # Build outputs
└── examples/           # Usage examples
```

### Building
```bash
# Development build
go build -o build/flyby ./cmd/flyby

# With make
make build

# Run tests
make test
```

### Architecture

FlyBy is built using:
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: TUI framework
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling
- **fly CLI**: Concourse operations via subprocess calls

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 License

MIT License

## 🚀 Roadmap

- [ ] Pipeline editing capabilities
- [ ] Resource version browsing
- [ ] Build log streaming
- [ ] Export/import configurations
- [ ] Custom themes
- [ ] Multi-target operations

---

**FlyBy** - Making Concourse CI management a breeze! 🌪️