# go-musthave-metrics-tpl


## pprof differences (result - base)
**alloc_space**:
File: server
Build ID: 89b61b0eb8b99322e4eba871d990d7ab65c4cc00
Type: alloc_space
Time: 2025-10-13 22:29:31 CEST
Showing nodes accounting for 452.35MB, 4.73% of 9557.15MB total
Dropped 118 nodes (cum <= 47.79MB)
      flat  flat%   sum%        cum   cum%
-2806.14MB 29.36% 29.36% -2774.10MB 29.03%  go.uber.org/zap/internal/stacktrace.Capture
-1079.34MB 11.29% 40.66% -1079.34MB 11.29%  go.uber.org/zap.(*SugaredLogger).sweetenFields
  965.27MB 10.10% 30.56%   965.27MB 10.10%  net/http.Header.Clone (inline)
  860.77MB  9.01% 21.55%   860.77MB  9.01%  net/textproto.MIMEHeader.Set (inline)
 -574.01MB  6.01% 27.55%  -574.01MB  6.01%  go.uber.org/zap/zapcore.(*sliceArrayEncoder).AppendString
  527.12MB  5.52% 22.04%   527.12MB  5.52%  net/textproto.readMIMEHeader
  383.48MB  4.01% 18.03%   383.48MB  4.01%  io.init.func1
  374.56MB  3.92% 14.11%  1770.13MB 18.52%  net/http.(*conn).readRequest
 -318.51MB  3.33% 17.44%  -318.51MB  3.33%  go.uber.org/zap/buffer.(*Buffer).String (inline)
 -273.01MB  2.86% 20.30%  -273.01MB  2.86%  time.Time.Format
  250.08MB  2.62% 17.68%   250.08MB  2.62%  net/http.(*Request).WithContext (inline)
  238.07MB  2.49% 15.19%  1149.25MB 12.03%  net/http.readRequest
  206.52MB  2.16% 13.03%   394.02MB  4.12%  internal/fmtsort.Sort
  187.50MB  1.96% 11.07%   187.50MB  1.96%  reflect.copyVal
  179.52MB  1.88%  9.19%   681.04MB  7.13%  fmt.Sprint
  159.36MB  1.67%  7.52%   159.36MB  1.67%  sync.(*Pool).pinSlow
  154.51MB  1.62%  5.90%   213.51MB  2.23%  net/http.readTransfer
  147.28MB  1.54%  4.36%   147.28MB  1.54%  bufio.NewWriterSize (inline)
  129.50MB  1.36%  3.01%   129.50MB  1.36%  io.LimitReader (inline)
     115MB  1.20%  1.80%   150.05MB  1.57%  fmt.Sprintf
   99.51MB  1.04%  0.76%    99.51MB  1.04%  net/url.parse
      92MB  0.96%   0.2% -2223.71MB 23.27%  github.com/devize-ed/yapracproj-metrics.git/internal/handler.(*Handler).NewRouter.MiddlewareLogging.func1.1
   87.50MB  0.92%  1.11%    87.50MB  0.92%  reflect.packEface
   68.01MB  0.71%  1.83%    68.01MB  0.71%  context.withCancel (inline)
   67.51MB  0.71%  2.53%    67.51MB  0.71%  net.(*conn).Read
   39.50MB  0.41%  2.95%  1868.09MB 19.55%  net/http.Error
   33.50MB  0.35%  3.30%   378.85MB  3.96%  net/http.(*conn).serve
   30.03MB  0.31%  3.61%    30.03MB  0.31%  go.uber.org/zap/internal/bufferpool.init.NewPool.func1
      30MB  0.31%  3.92%       30MB  0.31%  context.WithValue
      28MB  0.29%  4.22%       28MB  0.29%  net/textproto.(*Reader).ReadLine (inline)
  -24.50MB  0.26%  3.96%   -24.50MB  0.26%  time.Duration.String (inline)
      20MB  0.21%  4.17%       20MB  0.21%  fmt.(*buffer).writeByte (inline)
      19MB   0.2%  4.37%       19MB   0.2%  github.com/go-chi/chi.NewRouteContext
   16.51MB  0.17%  4.54%    16.51MB  0.17%  go.uber.org/zap/internal/stacktrace.init.func1
  -16.50MB  0.17%  4.37%   -16.50MB  0.17%  internal/sync.runtime_SemacquireMutex
   15.50MB  0.16%  4.53%    83.51MB  0.87%  context.WithCancel
      14MB  0.15%  4.68%       14MB  0.15%  go.uber.org/zap/zapcore.init.func2
   13.50MB  0.14%  4.82%    13.50MB  0.14%  sync.(*poolChain).pushHead
  -11.46MB  0.12%  4.70%   -18.42MB  0.19%  compress/flate.NewWriter (inline)
    6.50MB 0.068%  4.77%     6.50MB 0.068%  fmt.init.func1
    5.50MB 0.058%  4.83%     5.50MB 0.058%  go.uber.org/zap/zapcore.init.func4
   -4.61MB 0.048%  4.78%   -11.96MB  0.13%  runtime/pprof.(*profileBuilder).emitLocation
    3.50MB 0.037%  4.81%    25.55MB  0.27%  net/http.Header.sortedKeyValues
    3.50MB 0.037%  4.85%     3.50MB 0.037%  net/http.(*connReader).startBackgroundRead
   -3.07MB 0.032%  4.82%    -6.96MB 0.073%  compress/flate.(*compressor).init
      -3MB 0.031%  4.79%       -3MB 0.031%  go.uber.org/zap/zapcore.init.func1
   -2.50MB 0.026%  4.76%    -3.50MB 0.037%  internal/profile.(*profileMerger).mapSample
    2.50MB 0.026%  4.79%     2.50MB 0.026%  net/textproto.NewReader (inline)
   -2.31MB 0.024%  4.76%    -2.31MB 0.024%  runtime/pprof.StartCPUProfile
   -2.28MB 0.024%  4.74%    -2.28MB 0.024%  compress/flate.newDeflateFast (inline)
   -0.51MB 0.0054%  4.73%    -4.01MB 0.042%  internal/profile.Merge
         0     0%  4.73%   554.63MB  5.80%  bufio.(*Writer).Flush
         0     0%  4.73%   -18.42MB  0.19%  compress/gzip.(*Writer).Write
         0     0%  4.73%   501.53MB  5.25%  fmt.(*pp).doPrint
         0     0%  4.73%     3.50MB 0.037%  fmt.(*pp).free
         0     0%  4.73%   501.03MB  5.24%  fmt.(*pp).printArg
         0     0%  4.73%   502.03MB  5.25%  fmt.(*pp).printValue
         0     0%  4.73%   -10.52MB  0.11%  fmt.Fprint
         0     0%  4.73%     2.56MB 0.027%  fmt.Fprintln
         0     0%  4.73%    23.54MB  0.25%  fmt.newPrinter
         0     0%  4.73%  -114.71MB  1.20%  github.com/devize-ed/yapracproj-metrics.git/internal/handler.(*Handler).NewRouter.HashMiddleware.func2.1
         0     0%  4.73%   993.92MB 10.40%  github.com/devize-ed/yapracproj-metrics.git/internal/handler.(*Handler).NewRouter.MiddlewareGzip.func3.1
         0     0%  4.73%       19MB   0.2%  github.com/devize-ed/yapracproj-metrics.git/internal/handler.(*Handler).NewRouter.NewRouter.NewMux.func11
         0     0%  4.73%   965.27MB 10.10%  github.com/devize-ed/yapracproj-metrics.git/internal/handler/middleware.(*loggingResponseWriter).WriteHeader
         0     0%  4.73%  -448.51MB  4.69%  github.com/devize-ed/yapracproj-metrics.git/internal/logger.Initialize.TimeEncoderOfLayout.func1
         0     0%  4.73% -1909.08MB 19.98%  github.com/go-chi/chi.(*Mux).ServeHTTP
         0     0%  4.73%  1868.09MB 19.55%  github.com/go-chi/chi.(*Mux).routeHTTP
         0     0%  4.73%   993.92MB 10.40%  github.com/go-chi/chi/middleware.StripSlashes.func1
         0     0%  4.73% -2756.58MB 28.84%  go.uber.org/zap.(*Logger).Check (inline)
         0     0%  4.73% -2756.58MB 28.84%  go.uber.org/zap.(*Logger).check
         0     0%  4.73%   125.46MB  1.31%  go.uber.org/zap.(*SugaredLogger).Debug
         0     0%  4.73% -2108.26MB 22.06%  go.uber.org/zap.(*SugaredLogger).Debugf
         0     0%  4.73%    -2201MB 23.03%  go.uber.org/zap.(*SugaredLogger).Infow
         0     0%  4.73% -4183.80MB 43.78%  go.uber.org/zap.(*SugaredLogger).log
         0     0%  4.73%   831.09MB  8.70%  go.uber.org/zap.getMessage
         0     0%  4.73%    42.06MB  0.44%  go.uber.org/zap/buffer.Pool.Get
         0     0%  4.73%    30.03MB  0.31%  go.uber.org/zap/internal/bufferpool.init.NewPool.New[go.shape.*uint8].func2
         0     0%  4.73%   100.12MB  1.05%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0%  4.73%   -12.49MB  0.13%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Put (inline)
         0     0%  4.73%       -3MB 0.031%  go.uber.org/zap/internal/stacktrace.(*Stack).Free
         0     0%  4.73%    16.51MB  0.17%  go.uber.org/zap/internal/stacktrace.init.New[go.shape.*uint8].func2
         0     0%  4.73%    20.52MB  0.21%  go.uber.org/zap/zapcore.(*CheckedEntry).AddCore (inline)
         0     0%  4.73% -1178.96MB 12.34%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0%  4.73%    20.52MB  0.21%  go.uber.org/zap/zapcore.(*ioCore).Check
         0     0%  4.73% -1169.48MB 12.24%  go.uber.org/zap/zapcore.(*ioCore).Write
         0     0%  4.73%   -24.50MB  0.26%  go.uber.org/zap/zapcore.(*jsonEncoder).AddDuration
         0     0%  4.73%   -24.50MB  0.26%  go.uber.org/zap/zapcore.(*jsonEncoder).AppendDuration
         0     0%  4.73%   -19.54MB   0.2%  go.uber.org/zap/zapcore.(*jsonEncoder).Clone
         0     0%  4.73%   174.23MB  1.82%  go.uber.org/zap/zapcore.(*jsonEncoder).EncodeEntry
         0     0%  4.73%    68.11MB  0.71%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0%  4.73%   -16.50MB  0.17%  go.uber.org/zap/zapcore.(*lockedWriteSyncer).Write
         0     0%  4.73%    35.54MB  0.37%  go.uber.org/zap/zapcore.(*sampler).Check
         0     0%  4.73%  -195.50MB  2.05%  go.uber.org/zap/zapcore.CapitalLevelEncoder
         0     0%  4.73%     -310MB  3.24%  go.uber.org/zap/zapcore.EntryCaller.TrimmedPath
         0     0%  4.73%   -24.50MB  0.26%  go.uber.org/zap/zapcore.Field.AddTo
         0     0%  4.73%  -513.01MB  5.37%  go.uber.org/zap/zapcore.ShortCallerEncoder
         0     0%  4.73%   -24.50MB  0.26%  go.uber.org/zap/zapcore.StringDurationEncoder
         0     0%  4.73%   -24.50MB  0.26%  go.uber.org/zap/zapcore.addFields (inline)
         0     0%  4.73% -1327.71MB 13.89%  go.uber.org/zap/zapcore.consoleEncoder.EncodeEntry
         0     0%  4.73%   -44.04MB  0.46%  go.uber.org/zap/zapcore.consoleEncoder.writeContext
         0     0%  4.73%  -448.51MB  4.69%  go.uber.org/zap/zapcore.encodeTimeLayout
         0     0%  4.73%    20.52MB  0.21%  go.uber.org/zap/zapcore.getCheckedEntry
         0     0%  4.73%   -17.53MB  0.18%  go.uber.org/zap/zapcore.getSliceEncoder (inline)
         0     0%  4.73%       -3MB 0.031%  go.uber.org/zap/zapcore.init.New[go.shape.*uint8].func5
         0     0%  4.73%       14MB  0.15%  go.uber.org/zap/zapcore.init.New[go.shape.*uint8].func6
         0     0%  4.73%     5.50MB 0.058%  go.uber.org/zap/zapcore.init.New[go.shape.*uint8].func8
         0     0%  4.73%    -9.49MB 0.099%  go.uber.org/zap/zapcore.putCheckedEntry (inline)
         0     0%  4.73%     2.50MB 0.026%  go.uber.org/zap/zapcore.putJSONEncoder
         0     0%  4.73%    -2.50MB 0.026%  go.uber.org/zap/zapcore.putSliceEncoder (inline)
         0     0%  4.73%    -3.68MB 0.039%  internal/profile.(*Profile).Write
         0     0%  4.73%    -2.53MB 0.026%  internal/profile.Parse
         0     0%  4.73%   -16.50MB  0.17%  internal/sync.(*Mutex).Lock (inline)
         0     0%  4.73%   -16.50MB  0.17%  internal/sync.(*Mutex).lockSlow
         0     0%  4.73%       29MB   0.3%  io.(*LimitedReader).Read
         0     0%  4.73%   445.05MB  4.66%  io.Copy (inline)
         0     0%  4.73%   515.55MB  5.39%  io.CopyN
         0     0%  4.73%   445.05MB  4.66%  io.copyBuffer
         0     0%  4.73%   445.05MB  4.66%  io.discard.ReadFrom
         0     0%  4.73%   -28.78MB   0.3%  net/http.(*ServeMux).ServeHTTP
         0     0%  4.73%       29MB   0.3%  net/http.(*body).Read
         0     0%  4.73%       29MB   0.3%  net/http.(*body).readLocked
         0     0%  4.73%   554.63MB  5.80%  net/http.(*chunkWriter).Write
         0     0%  4.73%   -12.52MB  0.13%  net/http.(*chunkWriter).close
         0     0%  4.73%   542.11MB  5.67%  net/http.(*chunkWriter).writeHeader
         0     0%  4.73%    67.51MB  0.71%  net/http.(*connReader).backgroundRead
         0     0%  4.73%   965.27MB 10.10%  net/http.(*response).WriteHeader
         0     0%  4.73%   543.10MB  5.68%  net/http.(*response).finishRequest
         0     0%  4.73% -2252.49MB 23.57%  net/http.HandlerFunc.ServeHTTP (partial-inline)
         0     0%  4.73%   860.77MB  9.01%  net/http.Header.Set (inline)
         0     0%  4.73%    26.56MB  0.28%  net/http.Header.WriteSubset (inline)
         0     0%  4.73%    26.56MB  0.28%  net/http.Header.writeSubset
         0     0%  4.73%  1868.09MB 19.55%  net/http.NotFound
         0     0%  4.73%   160.81MB  1.68%  net/http.newBufioWriterSize
         0     0%  4.73%    14.03MB  0.15%  net/http.newTextprotoReader
         0     0%  4.73%    29.01MB   0.3%  net/http.putTextprotoReader
         0     0%  4.73% -1937.87MB 20.28%  net/http.serverHandler.ServeHTTP
         0     0%  4.73%   -26.47MB  0.28%  net/http/pprof.Index
         0     0%  4.73%    -2.31MB 0.024%  net/http/pprof.Profile
         0     0%  4.73%   -13.72MB  0.14%  net/http/pprof.collectProfile
         0     0%  4.73%   -26.47MB  0.28%  net/http/pprof.handler.ServeHTTP
         0     0%  4.73%   -21.41MB  0.22%  net/http/pprof.handler.serveDeltaProfile
         0     0%  4.73%   527.12MB  5.52%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0%  4.73%    99.51MB  1.04%  net/url.ParseRequestURI
         0     0%  4.73%       75MB  0.78%  reflect.(*MapIter).Key
         0     0%  4.73%   112.50MB  1.18%  reflect.(*MapIter).Value
         0     0%  4.73%    87.50MB  0.92%  reflect.Value.Interface (inline)
         0     0%  4.73%    87.50MB  0.92%  reflect.valueInterface
         0     0%  4.73%   -16.25MB  0.17%  runtime/pprof.(*Profile).WriteTo
         0     0%  4.73%   -13.46MB  0.14%  runtime/pprof.(*profileBuilder).appendLocsForStack
         0     0%  4.73%    -3.65MB 0.038%  runtime/pprof.(*profileBuilder).build
         0     0%  4.73%   -13.28MB  0.14%  runtime/pprof.(*profileBuilder).flush
         0     0%  4.73%    -5.94MB 0.062%  runtime/pprof.(*profileBuilder).pbSample
         0     0%  4.73%    -4.15MB 0.043%  runtime/pprof.profileWriter
         0     0%  4.73%   -16.25MB  0.17%  runtime/pprof.writeHeap
         0     0%  4.73%   -16.25MB  0.17%  runtime/pprof.writeHeapInternal
         0     0%  4.73%   -16.25MB  0.17%  runtime/pprof.writeHeapProto
         0     0%  4.73%   -16.50MB  0.17%  sync.(*Mutex).Lock (inline)
         0     0%  4.73%   623.36MB  6.52%  sync.(*Pool).Get
         0     0%  4.73%    20.02MB  0.21%  sync.(*Pool).Put
         0     0%  4.73%   159.36MB  1.67%  sync.(*Pool).pin