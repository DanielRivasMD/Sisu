# Sisu

[![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](LICENSE)

## Overview
`sisu`, build habits, finnish strong


## Features

- important features:
  - feedback / context
  - micro-reviews
  - day-off; if review average is low, suggest day-off
  - coach system

- command structure:
  - stats
  - graph
  - streak
  - cal

- tech stack:
  - cobra / viper
  - bubbletea
  - sqlx
  - go-cal
  - julia - unicodeplots

- database architecture:
  - tasks: track the high-level routines or goals, including tag categories
  - sessions: each time “track” a task, a session is logged
  - milestones: used for incentives, streaks, or mastery checkpoints
  - reviews: weekly or periodic reflections to prevent dropouts
  - coach: prewritten or dynamic messages triggered by milestones or streaks
  - calendar: for `cal` command—important dates, breaks, or annotations

## Quickstart
```
```

## Installation

### **Language-Specific**
| Language   | Command                                                                 |
|------------|-------------------------------------------------------------------------|
| **Go**     | `go install github.com/DanielRivasMD/Sisu@latest`                  |

### **Pre-built Binaries**
Download from [Releases](https://github.com/DanielRivasMD/Sisu/releases).

## Usage

```
```

## Example
```
```

## Configuration

## Development

Build from source
```
git clone https://github.com/DanielRivasMD/Sisu
cd Sisu
```

## Language-Specific Setup

| Language | Dev Dependencies | Hot Reload           |
|----------|------------------|----------------------|
| Go       | `go >= 1.21`     | `air` (live reload)  |

## License
Copyright (c) 2025

See the [LICENSE](LICENSE) file for license details.

