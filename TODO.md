- go test -count 12800 -v -failfast -run TestVar_Super_varWithInitThenMultipleDecleration/subctx_race
  - rare dead lock not detected by race condition

