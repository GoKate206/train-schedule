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
  < Coming soon! >

### Assumptions
- < Coming soon >


### Technology Used
- [Scribble](https://pkg.go.dev/github.com/nanobox-io/golang-scribble): Mock a database with JSON
- [Testify](https://pkg.go.dev/github.com/stretchr/testify): Provide tools to test code
