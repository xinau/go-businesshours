# go-businesshours

go-businesshours is a library for parsing business hours like `Mon-Fri 09:00-17:00 Europe/Berlin` and
checking they contain a `time.Time`.

It's possible to define business hours for a single day by only supplying a start day like `Mon 09:00-17:00 Europe/Berlin`.
In case no location was provided on the business hours UTC is assumed.

## Usage

Package documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/xinau/go-businesshours).

```go
bh, _ := businesshours.ParseBusinessHours("Mon-Fri 08:00-20:00 Europe/Berlin")
if bh.ContainsTime(time.Now()) {
	fmt.Printf("within business hours: %s\n", date, hours)
}
```

## Issues and Contributing

If you find an issue with this library, please report an issue. If
you'd like, I welcome any contributions. Fork this library and submit
a pull request.
