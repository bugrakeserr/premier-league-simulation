package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// premier league team with all the basic info we need
type PremierLeagueTeam struct {
	ID           int
	Name         string
	ShortName    string
	BaseStrength int
	Form         string
	Position     int
}

// main team struct that holds all the stats
type Team struct {
	Name            string
	Played          int
	Won             int
	Drawn           int
	Lost            int
	GoalsFor        int
	GoalsAgainst    int
	GoalDifference  int
	Points          int
	BaseStrength    int
	CurrentStrength int
	Form            []string // keeping track of last 5 games: "W", "D", "L"
}

// league structure that contains everything
type League struct {
	Teams    []*Team
	Week     int
	Fixtures [][]Match
}

// single match with all the details
type Match struct {
	HomeTeam  *Team
	AwayTeam  *Team
	HomeGoals int
	AwayGoals int
	IsPlayed  bool
	IsFixed   bool // whether user manually changed the result
	Week      int  // which week this match belongs to
}

// mock premier league teams with realistic strengths
func getMockPremierLeagueTeams() []PremierLeagueTeam {
	return []PremierLeagueTeam{
		{ID: 1, Name: "Manchester City", ShortName: "MCI", BaseStrength: 85, Form: "", Position: 1},
		{ID: 2, Name: "Arsenal", ShortName: "ARS", BaseStrength: 82, Form: "", Position: 2},
		{ID: 3, Name: "Liverpool", ShortName: "LIV", BaseStrength: 83, Form: "", Position: 3},
		{ID: 4, Name: "Manchester United", ShortName: "MUN", BaseStrength: 80, Form: "", Position: 4},
		{ID: 5, Name: "Tottenham", ShortName: "TOT", BaseStrength: 79, Form: "", Position: 5},
		{ID: 6, Name: "Newcastle", ShortName: "NEW", BaseStrength: 78, Form: "", Position: 6},
		{ID: 7, Name: "Chelsea", ShortName: "CHE", BaseStrength: 77, Form: "", Position: 7},
		{ID: 8, Name: "Aston Villa", ShortName: "AVL", BaseStrength: 76, Form: "", Position: 8},
		{ID: 9, Name: "Brighton", ShortName: "BHA", BaseStrength: 75, Form: "", Position: 9},
		{ID: 10, Name: "West Ham", ShortName: "WHU", BaseStrength: 74, Form: "", Position: 10},
	}
}

// randomly pick 4 teams from the list
func selectRandomTeams(teams []PremierLeagueTeam) []PremierLeagueTeam {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(teams), func(i, j int) {
		teams[i], teams[j] = teams[j], teams[i]
	})

	if len(teams) < 4 {
		return teams
	}
	return teams[:4]
}

// create a new league with 4 random teams
func NewLeague() *League {
	premierLeagueTeams := getMockPremierLeagueTeams()
	selectedTeams := selectRandomTeams(premierLeagueTeams)

	leagueTeams := make([]*Team, 4)
	for i, team := range selectedTeams {
		// convert form string to array if needed
		form := make([]string, 5)
		for j, result := range team.Form {
			form[j] = string(result)
		}

		leagueTeams[i] = &Team{
			Name:            team.Name,
			BaseStrength:    team.BaseStrength,
			CurrentStrength: team.BaseStrength,
			Form:            form,
		}
	}

	return &League{
		Teams: leagueTeams,
		Week:  0,
	}
}

// update team strength based on recent form
func (t *Team) updateTeamStrength() {
	// start with the base strength
	t.CurrentStrength = t.BaseStrength

	// calculate how form affects strength
	formMultiplier := 1.0
	for i, result := range t.Form {
		weight := float64(5-i) / 15.0 // recent games matter more
		switch result {
		case "W":
			formMultiplier += 0.05 * weight // wins boost strength
		case "D":
			// draws don't change anything
		case "L":
			formMultiplier -= 0.05 * weight // losses hurt strength
		}
	}

	// apply the form modifier
	t.CurrentStrength = int(float64(t.BaseStrength) * formMultiplier)

	// don't let it go too crazy - cap at ±15%
	minStrength := int(float64(t.BaseStrength) * 0.85)
	maxStrength := int(float64(t.BaseStrength) * 1.15)

	if t.CurrentStrength < minStrength {
		t.CurrentStrength = minStrength
	} else if t.CurrentStrength > maxStrength {
		t.CurrentStrength = maxStrength
	}
}

// UpdateTeamStats updates a team's statistics after a match
func (t *Team) UpdateTeamStats(goalsFor, goalsAgainst int) {
	t.GoalsFor += goalsFor
	t.GoalsAgainst += goalsAgainst
	t.GoalDifference = t.GoalsFor - t.GoalsAgainst

	if goalsFor > goalsAgainst {
		t.Won++
		t.Points += 3
		t.Form = append([]string{"W"}, t.Form[:4]...)
	} else if goalsFor == goalsAgainst {
		t.Drawn++
		t.Points += 1
		t.Form = append([]string{"D"}, t.Form[:4]...)
	} else {
		t.Lost++
		t.Form = append([]string{"L"}, t.Form[:4]...)
	}
	t.Played = t.Won + t.Drawn + t.Lost
	t.updateTeamStrength()
}

// reverse team stats when undoing a match result
func (t *Team) ReverseTeamStats(goalsFor, goalsAgainst int) {
	t.GoalsFor -= goalsFor
	t.GoalsAgainst -= goalsAgainst
	t.GoalDifference = t.GoalsFor - t.GoalsAgainst

	if goalsFor > goalsAgainst {
		t.Won--
		t.Points -= 3
	} else if goalsFor == goalsAgainst {
		t.Drawn--
		t.Points -= 1
	} else {
		t.Lost--
	}
	t.Played = t.Won + t.Drawn + t.Lost
}

// predict match result based on team strengths - this is where the magic happens
func predictMatchResult(team1, team2 *Team) (int, int) {
	// figure out total strength
	totalStrength := team1.CurrentStrength + team2.CurrentStrength

	// calculate team1's chance of winning
	team1Prob := float64(team1.CurrentStrength) / float64(totalStrength)

	// roll the dice
	rand.Seed(time.Now().UnixNano())
	r := rand.Float64()

	// decide the score based on probability
	var team1Goals, team2Goals int

	if r < team1Prob {
		// team 1 wins
		team1Goals = rand.Intn(3) + 1
		team2Goals = rand.Intn(team1Goals)
	} else if r < team1Prob+0.2 {
		// it's a draw
		team1Goals = rand.Intn(2)
		team2Goals = team1Goals
	} else {
		// team 2 wins
		team2Goals = rand.Intn(3) + 1
		team1Goals = rand.Intn(team2Goals)
	}

	return team1Goals, team2Goals
}

// generate fixtures for 18 weeks with 4 teams
func (l *League) generateFixtures() [][]Match {
	teams := l.Teams
	numTeams := len(teams)
	if numTeams != 4 {
		panic("This fixture generator is designed for exactly 4 teams.")
	}

	// set up team indices for round-robin
	indices := make([]int, numTeams)
	for i := range indices {
		indices[i] = i
	}

	var allWeeks [][]Match

	// repeat the double round-robin 3 times to get 18 weeks
	for repeat := 0; repeat < 3; repeat++ {
		// first double round-robin (home/away)
		for half := 0; half < 2; half++ {
			// reset indices for each double round-robin
			idx := make([]int, numTeams)
			copy(idx, indices)
			for round := 0; round < numTeams-1; round++ {
				var week []Match
				for i := 0; i < numTeams/2; i++ {
					home := idx[i]
					away := idx[numTeams-1-i]
					if half == 1 {
						home, away = away, home // swap home/away for second half
					}
					week = append(week, Match{
						HomeTeam: teams[home],
						AwayTeam: teams[away],
					})
				}
				allWeeks = append(allWeeks, week)
				// rotate indices for next round
				tmp := idx[1]
				copy(idx[1:numTeams-1], idx[2:])
				idx[numTeams-1] = tmp
			}
		}
	}

	return allWeeks
}

// SimulateNextWeek simulates the next week of matches
func (l *League) SimulateNextWeek() bool {
	if l.Week == 0 {
		l.Week = 1
	}

	// generate fixtures if this is the first week
	if l.Week == 1 {
		l.Fixtures = l.generateFixtures()
	}

	// check if we've played all weeks already
	if l.Week > len(l.Fixtures) {
		return false
	}

	// play this week's matches
	fmt.Printf("\nWeek %d Results:\n", l.Week)
	fmt.Println("----------------")

	for _, match := range l.Fixtures[l.Week-1] {
		homeGoals, awayGoals := predictMatchResult(match.HomeTeam, match.AwayTeam)
		fmt.Printf("%s %d - %d %s\n", match.HomeTeam.Name, homeGoals, awayGoals, match.AwayTeam.Name)

		match.HomeTeam.UpdateTeamStats(homeGoals, awayGoals)
		match.AwayTeam.UpdateTeamStats(awayGoals, homeGoals)
	}

	l.Week++
	return true
}

// print the league table nicely formatted
func (l *League) PrintLeagueTable() {
	// sort teams by points then goal difference
	sort.Slice(l.Teams, func(i, j int) bool {
		if l.Teams[i].Points != l.Teams[j].Points {
			return l.Teams[i].Points > l.Teams[j].Points
		}
		return l.Teams[i].GoalDifference > l.Teams[j].GoalDifference
	})

	// print the header
	fmt.Printf("\n%-20s %-8s %-8s %-8s %-8s %-8s %-8s %-8s %-8s %-8s\n",
		"Team", "Played", "Won", "Drawn", "Lost", "GF", "GA", "GD", "Points", "Strength")
	fmt.Println("----------------------------------------------------------------------------------------")

	// print each team's stats
	for _, team := range l.Teams {
		fmt.Printf("%-20s %-8d %-8d %-8d %-8d %-8d %-8d %-8d %-8d %-8d\n",
			team.Name, team.Played, team.Won, team.Drawn, team.Lost,
			team.GoalsFor, team.GoalsAgainst, team.GoalDifference, team.Points, team.CurrentStrength)
	}
}

// gui structure to handle the interface
type GUI struct {
	window         fyne.Window
	league         *League
	tableLabel     *widget.Label
	weekLabel      *widget.Label
	weekResults    *widget.Label // for week results
	allResults     *widget.Label // for season overview
	currentWeek    int           // which week we're currently viewing
	showAllResults bool          // whether to show the full season results
}

// create a new gui instance
func NewGUI() *GUI {
	myApp := app.New()
	window := myApp.NewWindow("Premier League Simulator")

	league := NewLeague()
	// set up fixtures right away
	league.Fixtures = league.generateFixtures()
	league.Week = 0

	gui := &GUI{
		window:         window,
		league:         league,
		tableLabel:     widget.NewLabel(""),
		weekLabel:      widget.NewLabel("Week 0"),
		weekResults:    widget.NewLabel(""),
		allResults:     widget.NewLabel(""),
		showAllResults: false,
	}

	gui.setupUI()
	return gui
}

// setupUI sets up the user interface
func (g *GUI) setupUI() {
	// set up standings with nice formatting
	standings := "Team                 P    W    D    L    GF   GA   GD   PTS\n"
	standings += "--------------------------------------------------------\n"
	for _, team := range g.league.Teams {
		standings += fmt.Sprintf("%-20s %3d  %3d  %3d  %3d  %3d  %3d  %3d  %3d\n",
			team.Name,
			team.Played,
			team.Won,
			team.Drawn,
			team.Lost,
			team.GoalsFor,
			team.GoalsAgainst,
			team.GoalDifference,
			team.Points)
	}

	// calculate starting probabilities based on base strength
	totalStrength := 0
	for _, team := range g.league.Teams {
		totalStrength += team.BaseStrength
	}
	initialProbs := make(map[string]float64)
	for _, team := range g.league.Teams {
		initialProbs[team.Name] = float64(team.BaseStrength) / float64(totalStrength) * 100.0
	}

	// sort teams by their probability
	type teamProb struct {
		name string
		prob float64
	}
	var probList []teamProb
	for name, prob := range initialProbs {
		probList = append(probList, teamProb{name, prob})
	}
	sort.Slice(probList, func(i, j int) bool {
		return probList[i].prob > probList[j].prob
	})

	// make the probability table with same width as standings
	probTable := "Championship Probability\n"
	probTable += "----------------------\n"
	for _, tp := range probList {
		probTable += fmt.Sprintf("%-20s %6.2f%%\n", tp.name, tp.prob)
	}

	standingsLabel := widget.NewLabelWithStyle(standings, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	probLabel := widget.NewLabelWithStyle(probTable, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	upcomingMatchesLabel := widget.NewLabelWithStyle(g.generateUpcomingMatchesTable(), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})

	// layout with proper spacing
	topRow := container.NewHBox(
		standingsLabel,
		widget.NewLabel("     "), // spacer to separate tables
		probLabel,
	)
	mainContent := container.NewVBox(
		topRow,
		widget.NewLabel(""), // vertical spacer
		upcomingMatchesLabel,
	)

	// button layout at the bottom
	simulateButton := widget.NewButton("Simulate Next Week", g.simulateNextWeek)
	playAllButton := widget.NewButton("Play All Remaining Weeks", g.simulateAllRemainingWeeks)
	buttonRow := container.NewHBox(
		simulateButton,
		widget.NewLabel("  "), // spacer between buttons
		playAllButton,
	)

	g.tableLabel.SetText("")
	g.window.SetContent(container.NewVBox(
		g.weekLabel,
		mainContent,
		widget.NewLabel(""),
		buttonRow,
	))

	g.window.Resize(fyne.NewSize(900, 700))
}

// simulateNextWeek simulates the next week of matches
func (g *GUI) simulateNextWeek() {
	if g.league.Week == 0 {
		g.league.Week = 1
		g.league.Fixtures = g.league.generateFixtures()
		// set week numbers for all matches
		for week := range g.league.Fixtures {
			for i := range g.league.Fixtures[week] {
				g.league.Fixtures[week][i].Week = week + 1
			}
		}
	}

	// check if we've reached the end of the season
	if g.league.Week > 18 {
		g.weekLabel.SetText("Season Completed!")
		g.weekResults.SetText("")
		g.refreshDisplay()
		return
	}

	// play the current week's matches
	weekMatches := g.league.Fixtures[g.league.Week-1]
	for i := range weekMatches {
		match := &g.league.Fixtures[g.league.Week-1][i]
		if !match.IsFixed {
			homeGoals, awayGoals := predictMatchResult(match.HomeTeam, match.AwayTeam)
			match.HomeGoals = homeGoals
			match.AwayGoals = awayGoals
			match.IsPlayed = true
		}
	}

	g.currentWeek = g.league.Week
	g.weekLabel.SetText(fmt.Sprintf("Week %d", g.league.Week))
	g.league.Week++

	// recalculate all the stats
	g.recalculateAllStats()

	// refresh the display
	g.refreshDisplay()
}

// recalculate all team stats from scratch
func (g *GUI) recalculateAllStats() {
	// reset all team stats first
	for _, team := range g.league.Teams {
		team.ResetTeamStats()
	}

	// go through all matches week by week
	for week := 0; week < g.league.Week; week++ {
		if week >= len(g.league.Fixtures) {
			break
		}
		for i := range g.league.Fixtures[week] {
			match := &g.league.Fixtures[week][i]
			if match.IsPlayed || match.IsFixed {
				match.HomeTeam.UpdateTeamStats(match.HomeGoals, match.AwayGoals)
				match.AwayTeam.UpdateTeamStats(match.AwayGoals, match.HomeGoals)
			}
		}
	}
}

// helper functions to make the display tables
func (g *GUI) generateStandingsTable() string {
	standings := "Team                 P    W    D    L    GF   GA   GD   PTS\n"
	standings += "--------------------------------------------------------\n"
	for _, team := range g.league.Teams {
		standings += fmt.Sprintf("%-20s %3d  %3d  %3d  %3d  %3d  %3d  %3d  %3d\n",
			team.Name,
			team.Played,
			team.Won,
			team.Drawn,
			team.Lost,
			team.GoalsFor,
			team.GoalsAgainst,
			team.GoalDifference,
			team.Points)
	}
	return standings
}

func (g *GUI) generateProbabilityTable() string {
	probs := g.league.ChampionshipProbabilities(10000)

	type teamProb struct {
		name string
		prob float64
	}
	var probList []teamProb
	for name, prob := range probs {
		probList = append(probList, teamProb{name, prob})
	}
	sort.Slice(probList, func(i, j int) bool {
		return probList[i].prob > probList[j].prob
	})

	probTable := "Championship Probability\n"
	probTable += "----------------------\n"
	for _, tp := range probList {
		probTable += fmt.Sprintf("%-20s %6.2f%%\n", tp.name, tp.prob)
	}
	return probTable
}

// editMatchResult opens a dialog for editing match result
func (g *GUI) editMatchResult(match *Match) {
	// create entry fields for the goals
	homeEntry := widget.NewEntry()
	homeEntry.SetText(fmt.Sprintf("%d", match.HomeGoals))
	homeEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		for _, r := range s {
			if r < '0' || r > '9' {
				return fmt.Errorf("only numbers allowed")
			}
		}
		val := 0
		if n, err := fmt.Sscanf(s, "%d", &val); err != nil || n != 1 {
			return fmt.Errorf("invalid number")
		}
		if val > 9 {
			return fmt.Errorf("maximum 9 goals allowed")
		}
		return nil
	}

	awayEntry := widget.NewEntry()
	awayEntry.SetText(fmt.Sprintf("%d", match.AwayGoals))
	awayEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		for _, r := range s {
			if r < '0' || r > '9' {
				return fmt.Errorf("only numbers allowed")
			}
		}
		val := 0
		if n, err := fmt.Sscanf(s, "%d", &val); err != nil || n != 1 {
			return fmt.Errorf("invalid number")
		}
		if val > 9 {
			return fmt.Errorf("maximum 9 goals allowed")
		}
		return nil
	}

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Week %d: %s vs %s", match.Week, match.HomeTeam.Name, match.AwayTeam.Name)),
		container.NewGridWithColumns(2,
			widget.NewLabel("Home Goals:"),
			homeEntry,
			widget.NewLabel("Away Goals:"),
			awayEntry,
		),
	)

	dialog := widget.NewModalPopUp(content, g.window.Canvas())

	buttons := container.NewHBox(
		widget.NewButton("Save", func() {
			homeGoals := 0
			if n, err := fmt.Sscanf(homeEntry.Text, "%d", &homeGoals); err != nil || n != 1 {
				return
			}

			awayGoals := 0
			if n, err := fmt.Sscanf(awayEntry.Text, "%d", &awayGoals); err != nil || n != 1 {
				return
			}

			// update the match result
			match.HomeGoals = homeGoals
			match.AwayGoals = awayGoals
			match.IsFixed = true
			match.IsPlayed = true

			// recalculate all stats
			g.recalculateAllStats()

			dialog.Hide()
			g.refreshDisplay()
		}),
		widget.NewButton("Cancel", func() {
			dialog.Hide()
		}),
	)

	dialog.Content = container.NewVBox(content, buttons)
	dialog.Resize(fyne.NewSize(300, 200))
	dialog.Show()
}

// refreshDisplay updates all display elements
func (g *GUI) refreshDisplay() {
	// sort teams by points and goal difference
	sort.Slice(g.league.Teams, func(i, j int) bool {
		if g.league.Teams[i].Points != g.league.Teams[j].Points {
			return g.league.Teams[i].Points > g.league.Teams[j].Points
		}
		return g.league.Teams[i].GoalDifference > g.league.Teams[j].GoalDifference
	})

	// create standings and probability tables
	standingsLabel := widget.NewLabelWithStyle(g.generateStandingsTable(), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
	probLabel := widget.NewLabelWithStyle(g.generateProbabilityTable(), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})

	topRow := container.NewHBox(
		standingsLabel,
		widget.NewLabel("     "),
		probLabel,
	)

	var mainContent fyne.CanvasObject

	if g.showAllResults {
		// show all season results when season is completed
		allResultsLabel := widget.NewLabelWithStyle(g.generateAllResultsTable(), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})

		// create scrollable container with explicit sizing
		scrollContainer := container.NewScroll(allResultsLabel)
		scrollContainer.SetMinSize(fyne.NewSize(1400, 500))
		scrollContainer.Resize(fyne.NewSize(1400, 500))

		mainContent = container.NewVBox(
			topRow,
			widget.NewLabel(""), // spacer
			scrollContainer,
		)
	} else {
		// show current week results and upcoming matches
		var resultButtons []fyne.CanvasObject
		if g.currentWeek > 0 && g.currentWeek <= len(g.league.Fixtures) {
			weekMatches := g.league.Fixtures[g.currentWeek-1]

			resultButtons = append(resultButtons, widget.NewLabel(fmt.Sprintf("Week %d Results:", g.currentWeek)))
			resultButtons = append(resultButtons, widget.NewLabel("----------------"))

			for i := range weekMatches {
				match := &g.league.Fixtures[g.currentWeek-1][i]
				resultText := fmt.Sprintf("%s %d - %d %s",
					match.HomeTeam.Name, match.HomeGoals,
					match.AwayGoals, match.AwayTeam.Name)

				btn := widget.NewButton(resultText, func() {
					g.editMatchResult(match)
				})
				if match.IsFixed {
					btn.Importance = widget.HighImportance
				}
				resultButtons = append(resultButtons, btn)
			}
		}

		resultsContainer := container.NewVBox(resultButtons...)

		// create upcoming matches table - show for week 0 through week 18
		var upcomingMatchesLabel *widget.Label
		if g.league.Week <= 18 {
			upcomingMatchesLabel = widget.NewLabelWithStyle(g.generateUpcomingMatchesTable(), fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
		} else {
			upcomingMatchesLabel = widget.NewLabel("")
		}

		mainContent = container.NewVBox(
			topRow,
			widget.NewLabel(""),
			resultsContainer,
			widget.NewLabel(""),
			upcomingMatchesLabel,
		)
	}

	var bottomContent fyne.CanvasObject
	if g.league.Week > 18 || g.showAllResults {
		if g.showAllResults {
			// show back to final week button when viewing all results
			backButton := widget.NewButton("Back to Final Week", func() {
				g.showAllResults = false
				g.refreshDisplay()
			})
			championLabel := widget.NewLabelWithStyle("🏆 Season Completed! 🏆", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			bottomContent = container.NewVBox(championLabel, backButton)
		} else {
			// show view all results button when season completed but not viewing all results
			viewAllButton := widget.NewButton("View All Season Results", func() {
				g.showAllResults = true
				g.refreshDisplay()
			})
			championLabel := widget.NewLabelWithStyle("🏆 Season Completed! 🏆", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
			bottomContent = container.NewVBox(championLabel, viewAllButton)
		}
	} else {
		// button layout for simulation
		simulateButton := widget.NewButton("Simulate Next Week", g.simulateNextWeek)
		playAllButton := widget.NewButton("Play All Remaining Weeks", g.simulateAllRemainingWeeks)
		buttonRow := container.NewHBox(
			simulateButton,
			widget.NewLabel("  "), // spacer
			playAllButton,
		)
		bottomContent = buttonRow
	}

	g.window.SetContent(container.NewVBox(
		g.weekLabel,
		mainContent,
		widget.NewLabel(""),
		bottomContent,
	))

	// adjust window size based on what we're showing
	if g.showAllResults {
		g.window.Resize(fyne.NewSize(1500, 900)) // extra wide for all results
	} else {
		g.window.Resize(fyne.NewSize(900, 700)) // normal size
	}

	g.window.Canvas().Refresh(g.window.Content())
}

// generateUpcomingMatchesTable creates a table showing next week's matches
func (g *GUI) generateUpcomingMatchesTable() string {
	// special case for week 0 to show week 1 matches
	if g.league.Week == 0 {
		if g.league.Fixtures == nil {
			g.league.Fixtures = g.league.generateFixtures()
		}
		var sb strings.Builder
		sb.WriteString("Upcoming Matches (Week 1)\n")
		sb.WriteString("----------------------------------------\n")

		for _, match := range g.league.Fixtures[0] {
			sb.WriteString(fmt.Sprintf("%-20s vs %-20s\n", match.HomeTeam.Name, match.AwayTeam.Name))
		}
		return sb.String()
	}

	currentWeek := g.league.Week - 1 // adjust for the actual current week
	if currentWeek >= len(g.league.Fixtures) || currentWeek >= 18 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Upcoming Matches (Week " + fmt.Sprintf("%d", currentWeek+1) + ")\n")
	sb.WriteString("----------------------------------------\n")

	nextWeekMatches := g.league.Fixtures[currentWeek]
	for _, match := range nextWeekMatches {
		sb.WriteString(fmt.Sprintf("%-20s vs %-20s\n", match.HomeTeam.Name, match.AwayTeam.Name))
	}

	return sb.String()
}

// monte carlo simulation for championship probability - this is the fun part
func (l *League) ChampionshipProbabilities(simulations int) map[string]float64 {
	counts := make(map[string]float64)
	numTeams := len(l.Teams)
	if numTeams == 0 {
		return counts
	}

	// if season is over, just figure out who won
	if l.Week > 18 {
		maxPoints := -1
		maxGoalDiff := -999
		for _, t := range l.Teams {
			if t.Points > maxPoints {
				maxPoints = t.Points
			}
		}
		// find team(s) with best goal difference among those with max points
		var champions []string
		for _, t := range l.Teams {
			if t.Points == maxPoints {
				if t.GoalDifference > maxGoalDiff {
					maxGoalDiff = t.GoalDifference
					champions = []string{t.Name} // new leader
				} else if t.GoalDifference == maxGoalDiff {
					champions = append(champions, t.Name)
				}
			}
		}
		// give them the probabilities
		for _, name := range champions {
			counts[name] = 100.0 / float64(len(champions))
		}
		// everyone else gets 0%
		for _, t := range l.Teams {
			if counts[t.Name] == 0 {
				counts[t.Name] = 0.0
			}
		}
		return counts
	}

	// check if someone has already won mathematically
	maxPoints := make(map[string]int)
	currentPoints := make(map[string]int)

	// get current points and max possible for each team
	for _, t := range l.Teams {
		currentPoints[t.Name] = t.Points
		remainingGames := 18 - t.Played                     // total games is 18
		maxPoints[t.Name] = t.Points + (remainingGames * 3) // max points from remaining
	}

	// check if leader has already won
	leader := l.Teams[0] // teams are sorted by points
	leaderUnreachable := true
	for _, t := range l.Teams[1:] {
		if maxPoints[t.Name] >= currentPoints[leader.Name] {
			leaderUnreachable = false
			break
		}
	}

	if leaderUnreachable {
		// leader has mathematically won
		for _, t := range l.Teams {
			if t.Name == leader.Name {
				counts[t.Name] = 100.0
			} else {
				counts[t.Name] = 0.0
			}
		}
		return counts
	}

	// if it's the start, base it on team strengths
	if l.Week == 0 {
		totalStrength := 0
		for _, t := range l.Teams {
			totalStrength += t.BaseStrength
		}
		for _, t := range l.Teams {
			counts[t.Name] = float64(t.BaseStrength) / float64(totalStrength) * 100.0
		}
		return counts
	}

	// otherwise run the monte carlo simulation
	validSimulations := 0
	for sim := 0; sim < simulations; sim++ {
		// make copies of all teams
		teamsCopy := make([]*Team, numTeams)
		for i, t := range l.Teams {
			formCopy := make([]string, len(t.Form))
			copy(formCopy, t.Form)
			teamsCopy[i] = &Team{
				Name:            t.Name,
				Played:          t.Played,
				Won:             t.Won,
				Drawn:           t.Drawn,
				Lost:            t.Lost,
				GoalsFor:        t.GoalsFor,
				GoalsAgainst:    t.GoalsAgainst,
				GoalDifference:  t.GoalDifference,
				Points:          t.Points,
				BaseStrength:    t.BaseStrength,
				CurrentStrength: t.CurrentStrength,
				Form:            formCopy,
			}
		}
		// simulate the rest of the season
		for w := l.Week - 1; w < len(l.Fixtures); w++ {
			for _, match := range l.Fixtures[w] {
				var home, away *Team
				for _, t := range teamsCopy {
					if t.Name == match.HomeTeam.Name {
						home = t
					}
					if t.Name == match.AwayTeam.Name {
						away = t
					}
				}
				if home == nil || away == nil {
					continue // skip bad matches
				}
				hg, ag := predictMatchResult(home, away)
				home.UpdateTeamStats(hg, ag)
				away.UpdateTeamStats(ag, hg)
			}
		}
		// find who won based on points and goal difference
		maxPoints := -1
		maxGoalDiff := -999
		for _, t := range teamsCopy {
			if t.Points > maxPoints {
				maxPoints = t.Points
			}
		}
		var champions []string
		for _, t := range teamsCopy {
			if t.Points == maxPoints {
				if t.GoalDifference > maxGoalDiff {
					maxGoalDiff = t.GoalDifference
					champions = []string{t.Name}
				} else if t.GoalDifference == maxGoalDiff {
					champions = append(champions, t.Name)
				}
			}
		}
		if len(champions) == 0 {
			continue // skip if something went wrong
		}
		for _, name := range champions {
			counts[name] += 1.0 / float64(len(champions))
		}
		validSimulations++
	}
	// turn counts into percentages
	if validSimulations > 0 {
		for name := range counts {
			counts[name] = counts[name] / float64(validSimulations) * 100.0
		}
	} else {
		// if no simulations worked, everyone gets 0%
		for _, t := range l.Teams {
			counts[t.Name] = 0.0
		}
	}
	return counts
}

// ResetTeamStats resets all team statistics to zero
func (t *Team) ResetTeamStats() {
	t.Played = 0
	t.Won = 0
	t.Drawn = 0
	t.Lost = 0
	t.GoalsFor = 0
	t.GoalsAgainst = 0
	t.GoalDifference = 0
	t.Points = 0
}

// simulate all remaining weeks until the season ends automatically
func (g *GUI) simulateAllRemainingWeeks() {
	if g.league.Week == 0 {
		g.league.Week = 1
		g.league.Fixtures = g.league.generateFixtures()
		// set week numbers for all matches
		for week := range g.league.Fixtures {
			for i := range g.league.Fixtures[week] {
				g.league.Fixtures[week][i].Week = week + 1
			}
		}
	}

	// use a timer to go week by week without freezing the ui
	g.simulateWeekByWeek()
}

// simulate one week at a time using timers so we can see the progression
func (g *GUI) simulateWeekByWeek() {
	if g.league.Week > 18 {
		// season is done - update on main thread but don't show all results yet
		fyne.Do(func() {
			g.currentWeek = 18
			g.weekLabel.SetText("Season Completed!")
			// keep showAllResults false so user sees final week first
			g.refreshDisplay()
		})
		return
	}

	// play the current week's matches
	if g.league.Week-1 < len(g.league.Fixtures) {
		weekMatches := g.league.Fixtures[g.league.Week-1]
		for i := range weekMatches {
			match := &g.league.Fixtures[g.league.Week-1][i]
			if !match.IsFixed {
				homeGoals, awayGoals := predictMatchResult(match.HomeTeam, match.AwayTeam)
				match.HomeGoals = homeGoals
				match.AwayGoals = awayGoals
				match.IsPlayed = true
			}
		}
	}

	g.currentWeek = g.league.Week
	g.league.Week++

	// recalculate all stats after each week
	g.recalculateAllStats()

	// update display on main thread
	fyne.Do(func() {
		g.weekLabel.SetText(fmt.Sprintf("Week %d", g.currentWeek))
		g.refreshDisplay()
	})

	// schedule the next week after a delay (500ms)
	time.AfterFunc(500*time.Millisecond, func() {
		g.simulateWeekByWeek()
	})
}

// generate a big table with all match results by week
func (g *GUI) generateAllResultsTable() string {
	if !g.showAllResults || len(g.league.Fixtures) == 0 {
		return ""
	}

	var results strings.Builder
	results.WriteString("ALL SEASON RESULTS\n")
	results.WriteString("==================================================\n\n")

	for week := 0; week < len(g.league.Fixtures) && week < 18; week++ {
		results.WriteString(fmt.Sprintf("Week %d:\n", week+1))
		results.WriteString("--------------------------------------------------\n")

		for _, match := range g.league.Fixtures[week] {
			if match.IsPlayed || match.IsFixed {
				// use max width for team names with clear spacing
				homeTeam := fmt.Sprintf("%-35s", match.HomeTeam.Name)
				awayTeam := fmt.Sprintf("%-35s", match.AwayTeam.Name)
				score := fmt.Sprintf("%d - %d", match.HomeGoals, match.AwayGoals)

				results.WriteString(fmt.Sprintf("%s  %s  %s", homeTeam, score, awayTeam))

				if match.IsFixed {
					results.WriteString("  (FIXED)")
				}
				results.WriteString("\n")
			}
		}
		results.WriteString("\n")
	}

	return results.String()
}
