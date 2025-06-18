# Project Snapshot

Project Snapshot is a Go-based application for managing and switching between project snapshots. It allows you to save, restore, and switch application states and window configurations efficiently.

## Features

- Save the current state of applications and their window configurations as snapshots.
- Restore applications and their windows from saved snapshots.
- Switch between different snapshots seamlessly.
- Quit unnecessary applications when switching snapshots.

## Requirements
- macOS (uses `osascript` for application management)
- yabai (for window management)

## Usage
## Save a Snapshot
Save the current state of applications and their windows:
```bash
projsnap take --name "SnapshotName"

# Optionally, quit applications after saving the snapshot
projsnap take --name "SnapshotName" --quit
```

## Restore a Snapshot
Restore applications and windows from a saved snapshot:
```bash
# restore = only open app(in snapshot)
projsnap restore --name "SnapshotName"
```

## Switch Snapshots
Switch between snapshots, closing unnecessary applications:
```bash
# switch = open app(in snapshot) + close apps(not in snapshot)
projsnap switch --name "SnapshotName"
```

## Remove Snapshots
Remove snapshots:
```bash
projsnap rm --name "SnapshotName"
```