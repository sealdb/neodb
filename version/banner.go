/*
 * NeoDB
 *
 * Copyright 2021-2030 The NeoDB Authors.
 * Code is licensed under the GPLv3.
 *
 */

package version


var banner = []string{
	// cyberlarge
	`
 __   _ _______  _____  ______  ______
 | \  | |______ |     | |     \ |_____]
 |  \_| |______ |_____| |_____/ |_____]
`,
	// doh
	`
NNNNNNNN        NNNNNNNN                                     DDDDDDDDDDDDD      BBBBBBBBBBBBBBBBB
N:::::::N       N::::::N                                     D::::::::::::DDD   B::::::::::::::::B
N::::::::N      N::::::N                                     D:::::::::::::::DD B::::::BBBBBB:::::B
N:::::::::N     N::::::N                                     DDD:::::DDDDD:::::DBB:::::B     B:::::B
N::::::::::N    N::::::N    eeeeeeeeeeee       ooooooooooo     D:::::D    D:::::D B::::B     B:::::B
N:::::::::::N   N::::::N  ee::::::::::::ee   oo:::::::::::oo   D:::::D     D:::::DB::::B     B:::::B
N:::::::N::::N  N::::::N e::::::eeeee:::::eeo:::::::::::::::o  D:::::D     D:::::DB::::BBBBBB:::::B
N::::::N N::::N N::::::Ne::::::e     e:::::eo:::::ooooo:::::o  D:::::D     D:::::DB:::::::::::::BB
N::::::N  N::::N:::::::Ne:::::::eeeee::::::eo::::o     o::::o  D:::::D     D:::::DB::::BBBBBB:::::B
N::::::N   N:::::::::::Ne:::::::::::::::::e o::::o     o::::o  D:::::D     D:::::DB::::B     B:::::B
N::::::N    N::::::::::Ne::::::eeeeeeeeeee  o::::o     o::::o  D:::::D     D:::::DB::::B     B:::::B
N::::::N     N:::::::::Ne:::::::e           o::::o     o::::o  D:::::D    D:::::D B::::B     B:::::B
N::::::N      N::::::::Ne::::::::e          o:::::ooooo:::::oDDD:::::DDDDD:::::DBB:::::BBBBBB::::::B
N::::::N       N:::::::N e::::::::eeeeeeee  o:::::::::::::::oD:::::::::::::::DD B:::::::::::::::::B
N::::::N        N::::::N  ee:::::::::::::e   oo:::::::::::oo D::::::::::::DDD   B::::::::::::::::B
NNNNNNNN         NNNNNNN    eeeeeeeeeeeeee     ooooooooooo   DDDDDDDDDDDDD      BBBBBBBBBBBBBBBBB
`,
}

func GetBanner() *string {
	return &banner[1]
}
