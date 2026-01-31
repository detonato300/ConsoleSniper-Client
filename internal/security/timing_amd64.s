//go:build amd64

TEXT ·rdtsc(SB), $0-8
    RDTSC
    SHLQ $32, DX
    ORQ DX, AX
    MOVQ AX, ret+0(FP)
    RET
