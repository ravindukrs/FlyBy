# FlyBy Usage Guide

## Getting Started

### Prerequisites

1. **Install fly CLI**: Download from [Concourse CI](https://concourse-ci.org/download.html)
2. **Configure targets**: Set up your Concourse targets using `fly login`
3. **Build FlyBy**: Run `make build` in the project directory

### Basic Usage

Start FlyBy by running:
```bash
./build/flyby
```

Or install it to your PATH:
```bash
make install
flyby
```

## Navigation Flow

FlyBy follows a hierarchical navigation structure:

```
Main Menu → Targets → Pipelines → Jobs/Resources → Builds
```

Each level provides specific functionality while maintaining easy navigation back to previous levels.

## Keyboard Controls

### Global Controls
- **Arrow Keys** or **j/k**: Navigate up/down in lists
- **Enter**: Select/activate current item
- **Esc**: Go back to previous view
- **q**: Quit the application
- **F5**: 🆕 **Refresh current view** (works in all views)
- **Ctrl+C**: Force quit

### Target Management
- **Enter**: Select target and view pipelines
- **a**: Add new target
- **d**: Delete target

### Pipeline Operations
- **Enter**: View jobs for selected pipeline
- **j**: View jobs for selected pipeline
- **r**: View resources for selected pipeline
- **p**: Pause/unpause pipeline
- **t**: Trigger first job in pipeline
- **F5**: Refresh pipeline list

### Job Management
- **Enter** or **t**: Trigger selected job
- **b**: 🆕 **View build history** for selected job
- **F5**: Refresh job list

### Resource Operations
- **Enter** or **c**: Check selected resource
- **F5**: Refresh resource list

### Build Operations 🆕
- **Enter**: **Rerun selected build** (with same inputs)
- **F5**: Refresh build list

## 🆕 Build Management Features

### Build History View

When you press **'b'** in the jobs view, you'll see:
- Complete build history for the selected job
- Build numbers (sequential: #1, #2, #11, etc.)
- Build status with color coding:
  - 🟢 **[SUCCEEDED]**: Green
  - 🔴 **[FAILED]**: Red  
  - 🟡 **[STARTED/PENDING]**: Yellow
- Build timing information (duration and relative time)
- Detailed build information panel

### Build Rerunning vs Job Triggering

**FlyBy provides two distinct operations:**

#### 🔄 Build Rerunning (Builds View)
- **Purpose**: Re-execute a specific build with **identical inputs**
- **Command**: `fly rerun-build --job pipeline/job --build X`
- **Result**: Creates sub-builds (e.g., #11.1, #11.2)
- **Use Case**: Retry failed builds, reproduce issues
- **Behavior**: Matches Concourse Web UI "re-run with same inputs"

#### ▶️ Job Triggering (Jobs View)  
- **Purpose**: Start a new build with **latest inputs**
- **Command**: `fly trigger-job -j pipeline/job`
- **Result**: Creates new build number (e.g., #12, #13)
- **Use Case**: Deploy latest changes, run with fresh resources

### Real-time Operation Feedback

All operations provide comprehensive feedback:

#### Success Messages ✅
- **Job Trigger**: `✓ Successfully triggered job pipeline/job: started pipeline/job #15`
- **Build Rerun**: `✓ Successfully reran build pipeline/job #11: started pipeline/job #11.1`
- **Resource Check**: `✓ Successfully checked resource: found new version`

#### Error Handling ❌
- Clear error messages with specific details
- Automatic error cleanup after 5 seconds
- Graceful handling of authentication issues

#### Loading States 🔄
- Loading indicators during operations
- Real-time status updates
- Non-blocking UI (can navigate away during operations)

## Authentication System

### Automatic Detection
FlyBy automatically detects authentication requirements:
- Monitors for "not authorized" errors
- Switches to authentication view seamlessly
- Preserves navigation context

### Interactive Authentication Process
1. **Detection**: System identifies auth requirement
2. **Auth Screen**: Shows target details and login prompt
3. **Browser Login**: Opens default browser for authentication
4. **Completion**: Automatic return to requested view
5. **Error Handling**: Clear retry options

### Authentication Controls
- **Enter/y**: Start authentication (opens browser)
- **n**: Cancel and return to previous view
- **Esc**: Cancel authentication

## 🆕 Refresh Functionality

Press **F5** in any view to refresh current data:

### Pipeline View
- Reloads all pipelines and their status
- Updates paused/unpaused states
- Refreshes pipeline metadata

### Jobs View  
- Updates job list and status
- Refreshes last build information
- Updates next build details

### Resources View
- Reloads resource list and status
- Updates "last checked" timestamps
- Refreshes resource versions

### Builds View
- Reloads complete build history
- Shows new builds after operations
- Updates build status and timing

## Target Management

### Adding Targets via FlyBy
1. Navigate to "Manage Targets" from main menu
2. Press **'a'** to add new target
3. Fill in required fields:
   - **Name**: Target identifier
   - **URL**: Concourse instance URL  
   - **Team**: Team name (default: main)
4. Use **Tab** to navigate between fields
5. Press **Enter** to save

### Adding Targets via CLI
```bash
fly -t my-target login -c https://ci.example.com -n team-name
```

FlyBy automatically detects and uses these targets.

### Target Operations
- **Select**: Choose active target for operations
- **Add**: Create new target configurations  
- **Delete**: Remove targets from configuration
- **Auto-detect**: Reads existing ~/.flyrc configuration

## Configuration

### ~/.flyrc Format
FlyBy reads your existing fly configuration:

```yaml
targets:
  production:
    name: production
    url: https://ci.example.com
    team: main
    token: your-token
    
  staging:
    name: staging  
    url: https://staging-ci.example.com
    team: development
    insecure: true
```

### No Additional Configuration Required
- Uses existing fly CLI setup
- Inherits authentication tokens
- Respects target-specific settings

## Status Indicators and Visual Cues

### Pipeline Status
- **[PAUSED]**: Pipeline is paused
- **[ARCHIVED]**: Pipeline is archived
- **Color coding**: Green for active, gray for paused

### Job Status
- **[SUCCEEDED]**: Last build succeeded (green)
- **[FAILED]**: Last build failed (red)
- **[STARTED]**: Currently running (yellow)
- **[PENDING]**: Queued for execution (yellow)

### Build Status
- **Build numbers**: Sequential (#1, #2, #11) 
- **Sub-builds**: Reruns shown as #11.1, #11.2
- **Timing**: Relative time (13min ago) and duration (3s)
- **Detailed info**: Full timestamps in info panel

### Resource Status  
- **Last checked**: Human-readable timestamps
- **Resource type**: git, s3-resource, docker-image, etc.
- **Check status**: Success/failure indicators

## Advanced Features

### Real-time Updates
- Live feedback during operations
- Non-blocking operations
- Automatic status refresh after actions

### Error Recovery
- Graceful error handling
- Automatic retry suggestions
- Context-aware error messages

### Navigation Memory
- Remembers selected items when navigating back
- Preserves scroll position
- Maintains operation context

## Keyboard Shortcuts Quick Reference

| View | Key | Action |
|------|-----|--------|
| **Global** | F5 | Refresh current view |
| | ↑/↓, j/k | Navigate |
| | Enter | Select/Execute |
| | Esc | Go back |
| | q | Quit |
| **Targets** | a | Add target |
| | d | Delete target |
| **Pipelines** | j | View jobs |
| | r | View resources |
| | p | Pause/unpause |
| | t | Trigger |
| **Jobs** | t | Trigger job |
| | b | View builds |
| **Resources** | c | Check resource |
| **Builds** | Enter | Rerun build |

## Troubleshooting

### Common Issues

**"fly CLI not found"**
- Install fly CLI: `brew install --cask fly`  
- Ensure it's in PATH: `which fly`
- Test: `fly --version`

**"No targets configured"**
- Configure targets: `fly -t target login -c URL`
- Check configuration: `cat ~/.flyrc`
- Add via FlyBy: Press 'a' in targets view

**"Authentication required"**
- Use built-in auth: Select target and follow prompts
- Manual login: `fly -t target login`
- Check token: `fly -t target status`

**"Empty lists"**
- Check permissions for selected team
- Verify target has pipelines/jobs/resources
- Use F5 to refresh view
- Ensure proper team membership

### Performance Issues

**Slow loading**
- Large number of pipelines/jobs can slow loading
- Use F5 refresh instead of navigating away/back
- Consider using specific teams for large instances

**Connection issues**  
- Check network connectivity to Concourse instance
- Verify URL in target configuration
- Check for VPN/proxy requirements

## Tips and Best Practices

### Efficient Navigation
1. **Use keyboard shortcuts**: j/k for navigation, F5 for refresh
2. **Learn the hierarchy**: Targets → Pipelines → Jobs → Builds
3. **Use refresh strategically**: F5 to update data without re-navigation

### Build Management
1. **Understand the difference**: Job triggering vs build rerunning
2. **Use build rerun for**: Failed builds, debugging, reproduction
3. **Use job trigger for**: New deployments, latest changes

### Resource Management
1. **Check resources manually**: Use Enter/c to force resource checks
2. **Monitor timestamps**: "Last checked" shows resource freshness
3. **Refresh after checks**: F5 to see updated check times

### Target Management
1. **Use descriptive names**: production, staging, dev-team
2. **Organize by environment**: Separate prod/non-prod targets
3. **Keep auth current**: Login regularly to maintain tokens

## Planned Enhancements

- **Build logs**: Stream build output in real-time
- **Pipeline editing**: Modify pipelines directly
- **Resource versions**: Browse resource version history
- **Multi-target operations**: Operate across multiple targets
- **Advanced search**: Filter pipelines, jobs, builds
- **Custom themes**: Personalize the interface
- **Export capabilities**: Save configurations and data

## Support and Feedback

For issues, feature requests, or contributions:
1. Check the project repository
2. Review existing issues and documentation  
3. Submit detailed bug reports with steps to reproduce
4. Suggest enhancements with use cases

Happy flying with FlyBy! ✈️