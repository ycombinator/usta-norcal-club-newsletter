## USTA NorCal Club Newsletter

This project provides a CLI tool to generate a newsletter for a tennis club participating in a USTA Norcal League. The newsletter consists of recent past results and upcoming fixtures.

## Installation

TODO: Download a release.

## Development

1. Download and install [Go](https://golang.org/).

2. Install project dependencies.
   ```
   cd usta-norcal-club-newsletter
   go get
   ```

3. Set environment variables.
   ```
   export UNCN_CLUB_ORG_ID=226
   export UNCN_PAST_DURATION_DAYS=7
   export UNCN_FUTURE_DURATION_DAYS=7
   ```

4. Run the tool.
   ```
   go run main.go
   ```
