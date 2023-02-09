package testcase

// EnvKeySeed is the environment variable key that will be checked for a pseudo random seed,
// which will be used to randomize the order of executions between test cases.
const EnvKeySeed = `TESTCASE_SEED`

// EnvKeyOrdering is the environment variable key that will be checked for testCase determine
// what order of execution should be used between test cases in a testing group.
// The default sorting behavior is pseudo random based on an the seed.
//
// Mods:
// - defined: execute testCase in the order which they are being defined
// - random: pseudo random based ordering between tests.
const EnvKeyOrdering = `TESTCASE_ORDERING`
