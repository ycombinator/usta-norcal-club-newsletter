package formatters

type RecentFormatter interface {
	FormatRecent(data *PreparedData, cfg Config) error
}

type UpcomingFormatter interface {
	FormatUpcoming(data *PreparedData, cfg Config) error
}
