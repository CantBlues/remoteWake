$env:GOMIPS="softfloat";
$env:CGO_ENABLED=0;
$env:GOOS="linux";
go build ./main.go;