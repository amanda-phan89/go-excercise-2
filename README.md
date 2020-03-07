## Content
A crawler demo for using
- goroutine
- waitgroup
- gorm with mysql driver

## Example
```sh
cp .env_example .env
# change database config in .env
# then run the following command
go run main.go https://www.thesaigontimes.vn/121623/Gia-dau-bien-dong---do-dau?.html
```

## TODO
- Break file main.go
- Apply database singleton