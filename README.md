
# Backend: a tech challenge with Go.
## Configuration

### Commands:
- Clone the repository - `git clone https://github.com/pkwarc/deviation-rest`
- `cd ./deviation-rest`
- (Optionally) run tests - `go test ./...`
- Set a port number - `export STD_DEV_PORT=7777`
- Build the image and run the container - `docker build --tag deviation-rest:latest . && docker run -p $STD_DEV_PORT:$STD_DEV_PORT -e STD_DEV_PORT=$STD_DEV_PORT deviation-rest:latest`
- The api random mean endpoint should be available at 
http://localhost:7777/random/mean?requests=10&length=10

---
