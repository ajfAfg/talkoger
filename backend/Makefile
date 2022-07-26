# HANDLERS := save_message fetch_message
HANDLERS := ${shell find . -name '*.go' -depth 1 -exec basename {} .go \;}

.PHONEY: ship
ship:
	for handler in ${HANDLERS}; do \
		GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" $${handler}.go ; \
		zip $${handler}.zip $${handler} ; \
		mv $${handler}.zip ~/Downloads ; \
		rm $${handler} ; \
	done
