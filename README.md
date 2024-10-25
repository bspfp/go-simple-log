# go-simple-log

go의 기본 패키지 log의 출력을 콘솔과 파일에 같이 하기 위한 패키지

(This is a package that provides a way to log output to both the console and a file using the standard Go log package.)

## 사용법 (Usage)

```go
package main

import (
    "gosimplelog"
)

func main() {
    // 20일 동안의 로그 파일을 유지하도록 설정
    logger, err := gosimplelog.InitLogFile("./log", "test.log", 20)

    // 파일 로테이션 사용 안함
    // logger, err := gosimplelog.InitLogFile("./log", "test.log", 0)

    // 기본 log 패키지를 통해 파일에 기록
    log.Println("Hello World")
}

```

```bash
$ git submodule add https://github.com/bspfp/go-simple-log.git ./pkg/gosimplelog
```
