# Running Tests

Before anything, you should clone the repository and install the dependencies like go because.. obviously it is a go test bed 

> Note : first go through [certificates](certs.md) configuration procedure before running the tests

two machines are recommended to run the tests because of the netem procedure will help to simulate the network conditions on server machine

## Server side
To run the tests for the project, use the following command:

```bash
go test -run Test_http3server -v ./.
```

This command will run http3server test in the current module and its subpackages. Make sure you have Go installed and your environment is properly set up before running the tests.

```bash
go test -run Test_http2server -v ./.
```

This command will run http2server test.

## Client side
To run the tests for the becnhmarks, use the following command:

go to [http_request_tester](./http_request_tester) folder

```bash
go test -run Test_http2_request -v ./.
```
This command will run http2_request test.

```bash
go test -run Test_http3_request -v ./.
```
This command will run http3_request test.

You can modify the number of requests in `*_test.go` file.

## Testing scenarios

so after the setup is working, you can emulate the network conditions with [netem](netem/tc.md) and run the tests.

## Results

So try yourself and see the results because the results can be different on different machines.