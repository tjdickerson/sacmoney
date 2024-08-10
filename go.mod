module ghitub.com/tjdickerson/sacmoney

go 1.22.5

require internal/sacdb v1.0.0

require github.com/mattn/go-sqlite3 v1.14.22 // indirect

replace internal/sacdb => ./internal/sacdb
