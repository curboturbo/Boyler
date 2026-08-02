package core

type Signal int

const (
    // 1. GracefulStop (SIGTERM)
    SignalGracefulStop Signal = iota

    // 2. ForceStop / Kill (SIGKILL)
    SignalForceStop

    // 3. ReloadConfig (SIGHUP)
    SignalReloadConfig

    // 4. Pause / Suspend (SIGSTOP)
    SignalPause

    // 5. Resume / Continue (SIGCONT)
    SignalResume

    // 6. GracefulDebug (SIGQUIT)
    SignalDumpCore
)

