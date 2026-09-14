<picture>
  <source media="(prefers-color-scheme: dark)" srcset="../.github/banner-minimal.svg">
  <source media="(prefers-color-scheme: light)" srcset="../.github/banner-minimal-light.svg">
  <img alt="minimal — a small CLI logger. A folded-paper m carries its ::, ?? and !! status markers." src="../.github/banner-minimal-light.svg">
</picture>

`minimal` is a small CLI-focused logger with distinct status markers, automatic terminal color detection and separate standard and error output streams.

```go
log := minimal.New()

_ = log.Infof("building %s\n", target)
_ = log.Stepln("compiling packages")
_ = log.Subln("generated configuration")
_ = log.Successln("build complete")
_ = log.Warnln("cache unavailable")
_ = log.Errorln("build failed")
```

`Info`, `Step`, `Success`, `Warn` and `Sub` write to stdout by default. `Error` writes to stderr. The `f` methods do not add a newline; use a newline in the format or use the corresponding `ln` method.

## Configuration

```go
log := minimal.New(
	minimal.WithTarget(output),
	minimal.WithErrorTarget(errorOutput),
	minimal.WithColor(minimal.ColorAuto),
)
```

`ColorAuto` emits color only when the selected stream is a color-capable terminal. `ColorAlways` forces ANSI output, while `ColorNever` strips ANSI sequences. `minimal.ToStdErr()` routes every message to stderr.

Every write method returns the underlying writer error. `MustFail` and `MustExit` are convenience methods for terminal error paths where no error can be returned to the caller.
