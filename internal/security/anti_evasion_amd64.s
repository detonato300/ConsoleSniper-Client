//go:build amd64

TEXT ·cpuid_check(SB), $0-1
    MOVL $1, AX
    CPUID
    SHRL $31, CX
    ANDL $1, CX
    MOVB CX, ret+0(FP)
    RET
