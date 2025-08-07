templ:
	cd backend && templ generate

run:
	sam local start-api --port 3001

build:
	sam build

all: templ build run
