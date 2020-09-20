# ormbench
sqlx vs ozzo-dbx vs gorm vs pgx benchmarks

## Results:
```
Benchmark/pgxpool_limit:5-4                 184643 ns/op
Benchmark/sqlx_limit:5-4                    226758 ns/op
Benchmark/gorm_v2_limit:5-4                 225265 ns/op
Benchmark/ozzo-dbx_limit:5-4                228467 ns/op
========================================================
Benchmark/pgxpool_limit:50-4                277449 ns/op
Benchmark/sqlx_limit:50-4                   295939 ns/op
Benchmark/gorm_v2_limit:50-4                358661 ns/op
Benchmark/ozzo-dbx_limit:50-4               289180 ns/op
========================================================
Benchmark/pgxpool_limit:100-4               306729 ns/op
Benchmark/sqlx_limit:100-4                  373172 ns/op
Benchmark/gorm_v2_limit:100-4               490190 ns/op
Benchmark/ozzo-dbx_limit:100-4              390490 ns/op
========================================================
Benchmark/pgxpool_limit:500-4               679745 ns/op
Benchmark/sqlx_limit:500-4                  824782 ns/op
Benchmark/gorm_v2_limit:500-4              1337712 ns/op
Benchmark/ozzo-dbx_limit:500-4              880810 ns/op
========================================================
Benchmark/pgxpool_limit:10000-4            8859699 ns/op
Benchmark/sqlx_limit:10000-4              11396879 ns/op
Benchmark/gorm_v2_limit:10000-4           19962606 ns/op
Benchmark/ozzo-dbx_limit:10000-4          11860572 ns/op
========================================================
```

## Colnclusion
* Up to 100 records (the most basic cases) we don't see significant difference between time per operation.
* On the big queries (special cases which mostly recomended execute throught Scan method) the winner of course pgxpool as expected.
* But to my big wonder the sqlx and ozzo-dbx on the big queries were almost equal. I didn't expect this. 
* And of course gorm on the big data query lose with almost twice defeat because of reflection usage. I strongly don't recommend using it for large slices. Use cursor istead.