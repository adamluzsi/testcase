#!/usr/bin/env bash
set -e

(
	type mockgen
	type mktemp
	type go
) 1>/dev/null

export PACKAGE="mocks"
export IN_PATH="int_test.go"
export OUT_PATH="MockTB_gen.go"

function main() {
	generateMock
	updateMock
}

function generateMock() {
	mockgen -source "${IN_PATH}" -destination "${OUT_PATH}" -package "${PACKAGE}"
}

function updateMock() {
	local out="$(mktemp)"
	while read -r -d $'\n' line || [[ -n ${line} ]]; do
		echo "${line}" >>"${out}"
		if [[ ${line} =~ type\ +MockTB\ +struct\ +\{ ]]; then
			echo "testing.TB" >>"${out}"
		fi
	done <"${OUT_PATH}"
	mv -f "${out}" "${OUT_PATH}"
	go fmt "${OUT_PATH}" 1>/dev/null
}

main "${@}"
