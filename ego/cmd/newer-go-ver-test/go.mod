module example.com/newer-go-ver-test

go 1.99

replace example.com/testmod => ./testmod

require example.com/testmod v0.0.0-00010101000000-000000000000
