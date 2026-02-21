# heartbeat

## DSL

```twf
activity ProcessLargeFile(fileId: string) -> (ProcessResult):
    file = download(fileId)
    for (chunk in file.chunks):
        process(chunk)
        heartbeat(progress: {current: chunk, total: len(file.chunks)})
    return ProcessResult{success: true}
```

Calling the activity with heartbeat and timeout options:

```twf
workflow ProcessFiles(fileId: string) -> (ProcessResult):
    activity ProcessLargeFile(fileId) -> result
        options:
            start_to_close_timeout: 2h
            heartbeat_timeout: 30s
    close complete(result)
```

## Go

```go
func ProcessLargeFile(ctx context.Context, fileId string) (ProcessResult, error) {
    file, err := download(ctx, fileId)
    if err != nil {
        return ProcessResult{}, err
    }
    for _, chunk := range file.Chunks {
        if err := process(ctx, chunk); err != nil {
            return ProcessResult{}, err
        }
        activity.RecordHeartbeat(ctx, map[string]interface{}{
            "current": chunk,
            "total":   len(file.Chunks),
        })
    }
    return ProcessResult{Success: true}, nil
}
```

## Notes

- `heartbeat(args)` → `activity.RecordHeartbeat(ctx, details...)` — activity-only, never in workflows
- Heartbeat details are arbitrary; use a struct or map
- Set `HeartbeatTimeout` in `ActivityOptions` on the calling side — see [options.md](./options.md)
- To resume from heartbeat progress: `activity.GetHeartbeatDetails(ctx, &lastProgress)` at activity start
- The plain function shown above becomes a struct method in practice — see [activity-def.md](./activity-def.md) notes
