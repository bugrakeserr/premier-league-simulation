# Premier League Simulator

A sophisticated Premier League simulator built with Go and Fyne GUI framework, featuring dynamic team strength calculations, Monte Carlo championship probability analysis, and persistent SQLite database storage.

## Features

### Core Simulation
- **4-Team League**: Randomly selects 4 teams from 10 Premier League teams for each simulation
- **Dynamic Team Strength**: Team performance adapts based on recent form (last 5 matches)
- **18-Week Season**: Complete round-robin tournament with multiple cycles
- **Realistic Match Simulation**: Score prediction based on team strengths and form

### Advanced Analytics
- **Monte Carlo Analysis**: 10,000-simulation championship probability calculations
- **Real-time Probability Updates**: Championship chances recalculated after each week
- **Form-based Adjustments**: Team strength varies ±15% based on recent results

### Interactive Features
- **Week-by-Week Simulation**: Step through the season one week at a time
- **Automated Season Play**: "Play All Remaining Weeks" with visual progression
- **Match Result Editing**: Manually override any match result
- **Comprehensive Results View**: Season-wide results display with scrollable interface

### Database Integration
- **SQLite Persistence**: Complete database schema with foreign key constraints
- **Data Integrity**: Comprehensive validation and error handling
- **Schema Design**: Teams, leagues, matches, and probability tracking tables

## Technical Stack

- **Go 1.24.3**: Core application language
- **Fyne v2.6.1**: Cross-platform GUI framework
- **SQLite3**: Embedded database with CGO integration
- **Database Driver**: github.com/mattn/go-sqlite3

## Quick Start

### Prerequisites
- Go 1.24.3 or higher
- CGO enabled (for SQLite integration)
- macOS/Linux development environment

### Installation & Running

```bash
# Clone or download the project
git clone <https://github.com/bugrakeserr/premier-league-simulation>
cd premier-league-simulator

# Install dependencies
go mod tidy

# Build and run
make run

# Or build separately
make build
./bin/premier-league-simulator
```

## Database Schema

The application uses a comprehensive SQLite schema with the following tables:

- **teams**: Team information and statistics
- **leagues**: League metadata and current state
- **matches**: Individual match results and details
- **league_teams**: Many-to-many relationship between leagues and teams
- **championship_probabilities**: Monte Carlo simulation results

See `database_schema.sql` for complete schema definition and example queries.

## User Interface

### Main View
- **League Table**: Real-time standings with points, goal difference, and form
- **Championship Probabilities**: Live-updated chances based on Monte Carlo analysis
- **Upcoming Matches**: Preview of next week's fixtures

### Season Simulation
- **Single Week**: Simulate one week at a time with immediate results
- **Full Season**: Automated progression through all remaining weeks (500ms intervals)
- **Results Editing**: Click any match result to manually override the score
- **Season Overview**: Comprehensive view of all match results by week

### Visual Features
- **Dynamic Window Sizing**: Automatically adjusts for different content views
- **Thread-safe Updates**: Non-blocking simulation with proper GUI thread handling
- **Responsive Layout**: Optimized for different screen sizes and content types

## Build System

```bash
# Available make commands
make run          # Build and run the application
make build        # Build executable to bin/ directory
make clean        # Remove build artifacts
make deps         # Install/update dependencies
```

## Project Structure

```
├── main.go                    # Application entry point and database initialization
├── simulation.go              # Core simulation logic and GUI implementation
├── database.go               # Database operations and schema management
├── database_schema.sql       # Complete database schema and example queries
├── go.mod                    # Go module dependencies
├── go.sum                    # Dependency checksums
├── Makefile                  # Build automation
└── README.md                 # Project documentation
```

## Key Algorithms

### Team Strength Calculation
- Base strength from Premier League team ratings (74-85 range)
- Form multiplier based on last 5 results (±5% per win/loss)
- Capped at ±15% of base strength for realistic variance

### Match Prediction
- Probability-based outcome determination
- Goal calculation based on team strength ratios
- Random variation for realistic unpredictability

### Championship Probability
- 10,000-iteration Monte Carlo simulation
- Mathematical championship detection for early season completion
- Real-time recalculation after each week's results

## Development Notes

### Threading Model
- Main simulation runs on background threads
- GUI updates use `fyne.Do()` for thread safety
- Non-blocking timers for automated season progression

### Database Considerations
- CGO compilation required for SQLite3 driver
- Foreign key constraints enabled for data integrity
- Transaction-based operations for consistency

### Platform Compatibility
- Primary development and testing on macOS
- Cross-platform compatible via Fyne framework
- SQLite provides universal database compatibility

## Future Enhancements

- Player-level statistics and transfers
- Historical season tracking and comparison
- Export functionality for league results
- Custom team strength configuration
- Multi-league support with promotion/relegation

---

**Built with Go and Fyne for cross-platform desktop simulation gaming.** 
