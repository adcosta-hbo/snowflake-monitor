# asyncBufferWriteSyncer

> It's a BufferWriteSyncer that flushes asynchronously.

For logging, we want to buffer the log messages.  If logging is not buffered and done synchronously, then the program
will spend a lot of CPU on io.Write() (ie.. either stdout, or os.File.Write).  To reduce the number calls to io.Write(),
asyncBufferWriteSyncer will buffer the logs and call io.Write asynchronously when either
1) when flushInterval is expired, or 2) the accumulated buffer has exceeded the max bufferSize


Note:  **Because messages are buffered, please don't foget to flush the buffer before exiting the application**

### Usage

```
import (
  "os"
  "time"

  "github.com/HBOCodeLabs/hurley-kit/llog/buffered"
)

func main() {
    ws := buffered.NewAsyncBufferWriteSyncer(os.Stdout, 1024, time.Millisecond * 100)

    // Do NOT forget to call Sync() to flush the buffer when exiting the application
    defer ws.Sync()

    ws.Write([]byte("logging stuff"))
    // At this point, the log is not sent to stdout
    time.Sleep(time.Second)
    // At this point, the log is sent to stdout
}
```

