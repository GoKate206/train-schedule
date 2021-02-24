Train Schedule Program

## Requirements
- Customers should be able to upload a CSV of train schedules that include
  - Stop Id
  - Route
  - Train number ( 4 character alphanumeric )
  - Time arriving
- Customers can request a schedule from a given time
  - A schedule will be returned if there are two or more coming at that time
  - If there are no more trains coming at the end of the day show the first trains from the next day
  - If there are one or fewer trains arriving no times will be shown

### How to run this program
- Build the project and dependencies by running `go mod init src/github.com/GoKate206` ( Make sure that there is no leading slash at the end of `GoKate206`)
- Navigate to `src/github.com/GoKate206`
  - Run the tests `go test *.go -v`
  - Run the program `go run main.go`

### Assumptions
- Train schedules will only be returned if there are 2 or more trains coming at requested time
- Return value is a slice of struct with schedule details
- There is no given range from requested time ( ex: User wants to see a bus at 3:30 they will not see buses that come at 3:31 )
- Time is in military for comparisons


### Technology Used
- [Scribble](https://pkg.go.dev/github.com/nanobox-io/golang-scribble): Mock a database with JSON
- [Testify](https://pkg.go.dev/github.com/stretchr/testify): Provide tools for testing
