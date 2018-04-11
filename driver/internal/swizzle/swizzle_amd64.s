// Copyright 2018 (as). All rights reserved. Added bgra32 for CPUs
// with AVX2.

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

DATA ·AVX2_swizzletab<>+0x00(SB)/8, $0x0704050603000102
DATA ·AVX2_swizzletab<>+0x08(SB)/8, $0x0f0c0d0e0b08090a
DATA ·AVX2_swizzletab<>+0x10(SB)/8, $0x1714151613101112
DATA ·AVX2_swizzletab<>+0x18(SB)/8, $0x0f0c0d0e0b08090a
GLOBL ·AVX2_swizzletab<>(SB), (NOPTR+RODATA), $32

// func haveSSSE3() bool
TEXT ·haveSSSE3(SB),NOSPLIT,$0
	MOVQ	$1, AX
	CPUID
	SHRQ	$9, CX
	ANDQ	$1, CX
	MOVB	CX, ret+0(FP)
	RET

// func haveAVX2() bool
TEXT ·haveAVX2(SB),NOSPLIT,$0
	MOVQ	$7, AX
	MOVQ	$0, CX
	CPUID
	SHRQ	$5, BX
	ANDQ	$1, BX
	MOVB	BX, ret+0(FP)
	RET

// func bgra256sd(p, q []byte)
TEXT ·bgra256sd(SB),NOSPLIT,$0-24
	MOVQ	p+0(FP), SI
	MOVQ	len+8(FP), CX
	MOVQ	q+24(FP), DI
	
	VMOVDQU ·AVX2_swizzletab<>(SB), Y0
	ADDQ SI, CX
	MOVQ CX, AX	// AX contains the final end offset
	ANDB	$0, CX
	
loop:
	CMPQ CX, SI
	JLE loop32
	VMOVDQU 	(0*32)(SI),Y1 
	VMOVDQU 	(1*32)(SI),Y2 
	VMOVDQU 	(2*32)(SI),Y3 
	VMOVDQU 	(3*32)(SI),Y4 
	VMOVDQU 	(4*32)(SI),Y5 
	VMOVDQU 	(5*32)(SI),Y6 
	VMOVDQU 	(6*32)(SI),Y7 
	VMOVDQU 	(7*32)(SI),Y8 
	VPSHUFB Y0, Y1,  Y1
	VPSHUFB Y0, Y2,  Y2
	VPSHUFB Y0, Y3,  Y3
	VPSHUFB Y0, Y4,  Y4
	VPSHUFB Y0, Y5,  Y5
	VPSHUFB Y0, Y6,  Y6
	VPSHUFB Y0, Y7,  Y7
	VPSHUFB Y0, Y8,  Y8
//	VPSHUFB Y0, Y1,  Y1
//	VPSHUFB Y0, Y5,  Y5
//	VPSHUFB Y0, Y2,  Y2
//	VPSHUFB Y0, Y8,  Y8
//	VPSHUFB Y0, Y3,  Y3
//	VPSHUFB Y0, Y7,  Y7
//	VPSHUFB Y0, Y4,  Y4
//	VPSHUFB Y0, Y6,  Y6
	VMOVDQU	Y1, (0*32)(DI)
	VMOVDQU	Y2, (1*32)(DI)
	VMOVDQU	Y3, (2*32)(DI)
	VMOVDQU	Y4, (3*32)(DI)
	VMOVDQU	Y5, (4*32)(DI)
	VMOVDQU	Y6, (5*32)(DI)
	VMOVDQU	Y7, (6*32)(DI)
	VMOVDQU	Y8, (7*32)(DI)
	ADDQ	$256, SI
	ADDQ	$256, DI
	JMP	loop

loop32:
	CMPQ AX, SI
	JEQ done
	VMOVDQU 	(0*32)(SI),Y1 
	VPSHUFB Y0, Y1,  Y1
	VMOVDQU	Y1, (0*32)(DI)
	ADDQ	$4, SI
	ADDQ	$4, DI
	JMP	loop32
	
done:
	RET

// func bgra256(p []byte)
TEXT ·bgra256(SB),NOSPLIT,$0-24
	MOVQ	p+0(FP), SI
	MOVQ	len+8(FP), DI

	// Sanity check that len is a multiple of 64.
	MOVQ	DI, AX
	ANDQ	$63, AX
	JNZ	done

	VMOVDQU ·AVX2_swizzletab<>(SB), Y0
	ADDQ	SI, DI
loop:
	CMPQ	SI, DI
	JEQ	done
	
	VMOVDQU 	(0*32)(SI),Y1 
	VMOVDQU 	(1*32)(SI),Y2 
	VMOVDQU 	(2*32)(SI),Y3 
	VMOVDQU 	(3*32)(SI),Y4 
	VMOVDQU 	(4*32)(SI),Y5 
	VMOVDQU 	(5*32)(SI),Y6 
	VMOVDQU 	(6*32)(SI),Y7 
	VMOVDQU 	(7*32)(SI),Y8 
	VPSHUFB Y0, Y1,  Y1
	VPSHUFB Y0, Y5,  Y5
	VPSHUFB Y0, Y2,  Y2
	VPSHUFB Y0, Y8,  Y8
	VPSHUFB Y0, Y3,  Y3
	VPSHUFB Y0, Y7,  Y7
	VPSHUFB Y0, Y4,  Y4
	VPSHUFB Y0, Y6,  Y6
	VMOVDQU	Y1, (0*32)(SI)
	VMOVDQU	Y2, (1*32)(SI)
	VMOVDQU	Y3, (2*32)(SI)
	VMOVDQU	Y4, (3*32)(SI)
	VMOVDQU	Y5, (4*32)(SI)
	VMOVDQU	Y6, (5*32)(SI)
	VMOVDQU	Y7, (6*32)(SI)
	VMOVDQU	Y8, (7*32)(SI)

	ADDQ	$256, SI
	JMP	loop
done:
	RET

// func bgra64(p []byte)
TEXT ·bgra64(SB),NOSPLIT,$0-24
	MOVQ	p+0(FP), SI
	MOVQ	len+8(FP), DI


	// Sanity check that len is a multiple of 64.
	MOVQ	DI, AX
	ANDQ	$63, AX
	JNZ	done


	VMOVDQU ·AVX2_swizzletab<>(SB), Y0
	ADDQ	SI, DI
loop:
	CMPQ	SI, DI
	JEQ	done
	
	VMOVDQU 	(0*32)(SI),Y1 
	VMOVDQU 	(1*32)(SI),Y2 
	VPSHUFB Y0, Y1,  Y1
	VPSHUFB Y0, Y2,  Y2
	VMOVDQU	Y1, (0*32)(SI)
	VMOVDQU	Y2, (1*32)(SI)

	ADDQ	$64, SI
	JMP	loop
done:
	RET

// func bgra32(p []byte)
TEXT ·bgra32(SB),NOSPLIT,$0-24
	MOVQ	p+0(FP), SI
	MOVQ	len+8(FP), DI

	// Sanity check that len is a multiple of 32.
	MOVQ	DI, AX
	ANDQ	$31, AX
	JNZ	done

	VMOVDQU ·AVX2_swizzletab<>(SB), Y0
	ADDQ	SI, DI
loop:
	CMPQ	SI, DI
	JEQ	done
	
	VMOVDQU 	(SI),Y1 
	VPSHUFB Y0, Y1,  Y1
	VMOVDQU	Y1, (SI)

	ADDQ	$32, SI
	JMP	loop
done:
	RET

// func bgra16(p []byte)
TEXT ·bgra16(SB),NOSPLIT,$0-24
	MOVQ	p+0(FP), SI
	MOVQ	len+8(FP), DI

	// Sanity check that len is a multiple of 16.
	MOVQ	DI, AX
	ANDQ	$15, AX
	JNZ	done

	// Make the shuffle control mask (16-byte register X0) look like this,
	// where the low order byte comes first:
	//
	// 02 01 00 03  06 05 04 07  0a 09 08 0b  0e 0d 0c 0f
	//
	// Load the bottom 8 bytes into X0, the top into X1, then interleave them
	// into X0.
	MOVQ	$0x0704050603000102, AX
	MOVQ	AX, X0
	MOVQ	$0x0f0c0d0e0b08090a, AX
	MOVQ	AX, X1
	PUNPCKLQDQ	X1, X0

	ADDQ	SI, DI
loop:
	CMPQ	SI, DI
	JEQ	done

	MOVOU	(SI), X1
	PSHUFB	X0, X1
	MOVOU	X1, (SI)

	ADDQ	$16, SI
	JMP	loop
done:
	RET

// func bgra4(p []byte)
TEXT ·bgra4(SB),NOSPLIT,$0-24
	MOVQ	p+0(FP), SI
	MOVQ	len+8(FP), DI

	ADDQ	SI, DI
loop:
	CMPQ	SI, DI
	JEQ	done

	MOVB	0(SI), AX
	MOVB	2(SI), BX
	MOVB	BX, 0(SI)
	MOVB	AX, 2(SI)

	ADDQ	$4, SI
	JMP	loop
done:
	RET
