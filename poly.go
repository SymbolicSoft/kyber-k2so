/* SPDX-FileCopyrightText: © 2020-2021 Nadim Kobeissi <nadim@symbolic.software>
 * SPDX-License-Identifier: MIT */

package kyberk2so

type poly struct {
	coeffs []int16
}

type polyvec struct {
	vec []poly
}

func polyNew() poly {
	var p poly
	p.coeffs = make([]int16, paramsN)
	return p
}

func polyvecNew(paramsK int) polyvec {
	var pv polyvec
	pv.vec = make([]poly, paramsK)
	for i := 0; i < paramsK; i++ {
		pv.vec[i] = polyNew()
	}
	return pv
}

func polyCompress(a poly, paramsK int) []byte {
	t := make([]byte, 8)
	a = polyCSubQ(a)
	rr := 0
	switch paramsK {
	case 2:
		r := make([]byte, paramsPolyCompressedBytesK2)
		for i := 0; i < paramsN/8; i++ {
			for j := 0; j < 8; j++ {
				t[j] = byte(((uint16(a.coeffs[8*i+j])<<3)+uint16(paramsQ)/2)/uint16(paramsQ)) & 7
			}
			r[rr+0] = (t[0] >> 0) | (t[1] << 3) | (t[2] << 6)
			r[rr+1] = (t[2] >> 2) | (t[3] << 1) | (t[4] << 4) | (t[5] << 7)
			r[rr+2] = (t[5] >> 1) | (t[6] << 2) | (t[7] << 5)
			rr = rr + 3
		}
		return r
	case 3:
		r := make([]byte, paramsPolyCompressedBytesK3)
		for i := 0; i < paramsN/8; i++ {
			for j := 0; j < 8; j++ {
				t[j] = byte(((uint16(a.coeffs[8*i+j])<<4)+uint16(paramsQ/2))/uint16(paramsQ)) & 15
			}
			r[rr+0] = t[0] | (t[1] << 4)
			r[rr+1] = t[2] | (t[3] << 4)
			r[rr+2] = t[4] | (t[5] << 4)
			r[rr+3] = t[6] | (t[7] << 4)
			rr = rr + 4
		}
		return r
	default:
		r := make([]byte, paramsPolyCompressedBytesK4)
		for i := 0; i < paramsN/8; i++ {
			for j := 0; j < 8; j++ {
				t[j] = byte(((uint32(a.coeffs[8*i+j])<<5)+uint32(paramsQ/2))/uint32(paramsQ)) & 31
			}
			r[rr+0] = (t[0] >> 0) | (t[1] << 5)
			r[rr+1] = (t[1] >> 3) | (t[2] << 2) | (t[3] << 7)
			r[rr+2] = (t[3] >> 1) | (t[4] << 4)
			r[rr+3] = (t[4] >> 4) | (t[5] << 1) | (t[6] << 6)
			r[rr+4] = (t[6] >> 2) | (t[7] << 3)
			rr = rr + 5
		}
		return r
	}
}

func polyDecompress(a []byte, paramsK int) poly {
	r := polyNew()
	t := make([]byte, 8)
	aa := 0
	switch paramsK {
	case 2:
		for i := 0; i < paramsN/8; i++ {
			t[0] = (a[aa+0] >> 0)
			t[1] = (a[aa+0] >> 3)
			t[2] = (a[aa+0] >> 6) | (a[aa+1] << 2)
			t[3] = (a[aa+1] >> 1)
			t[4] = (a[aa+1] >> 4)
			t[5] = (a[aa+1] >> 7) | (a[aa+2] << 1)
			t[6] = (a[aa+2] >> 2)
			t[7] = (a[aa+2] >> 5)
			aa = aa + 3
			for j := 0; j < 8; j++ {
				r.coeffs[8*i+j] = int16(((uint32(t[j]&7) * uint32(paramsQ)) + 4) >> 3)
			}
		}
	case 3:
		for i := 0; i < paramsN/2; i++ {
			r.coeffs[2*i+0] = int16(((uint16(a[aa]&15) * uint16(paramsQ)) + 8) >> 4)
			r.coeffs[2*i+1] = int16(((uint16(a[aa]>>4) * uint16(paramsQ)) + 8) >> 4)
			aa = aa + 1
		}
	case 4:
		for i := 0; i < paramsN/8; i++ {
			t[0] = (a[aa+0] >> 0)
			t[1] = (a[aa+0] >> 5) | (a[aa+1] << 3)
			t[2] = (a[aa+1] >> 2)
			t[3] = (a[aa+1] >> 7) | (a[aa+2] << 1)
			t[4] = (a[aa+2] >> 4) | (a[aa+3] << 4)
			t[5] = (a[aa+3] >> 1)
			t[6] = (a[aa+3] >> 6) | (a[aa+4] << 2)
			t[7] = (a[aa+4] >> 3)
			aa = aa + 5
			for j := 0; j < 8; j++ {
				r.coeffs[8*i+j] = int16(((uint32(t[j]&31) * uint32(paramsQ)) + 16) >> 5)
			}
		}
	}
	return r
}

func polyToBytes(a poly) []byte {
	var t0, t1 uint16
	r := make([]byte, paramsPolyBytes)
	a = polyCSubQ(a)
	for i := 0; i < paramsN/2; i++ {
		t0 = uint16(a.coeffs[2*i])
		t1 = uint16(a.coeffs[2*i+1])
		r[3*i+0] = byte(t0 >> 0)
		r[3*i+1] = byte(t0>>8) | byte(t1<<4)
		r[3*i+2] = byte(t1 >> 4)
	}
	return r
}

func polyFromBytes(a []byte) poly {
	r := polyNew()
	for i := 0; i < paramsN/2; i++ {
		r.coeffs[2*i] = int16(((uint16(a[3*i+0]) >> 0) | (uint16(a[3*i+1]) << 8)) & 0xFFF)
		r.coeffs[2*i+1] = int16(((uint16(a[3*i+1]) >> 4) | (uint16(a[3*i+2]) << 4)) & 0xFFF)
	}
	return r
}

func polyFromMsg(msg []byte) poly {
	r := polyNew()
	var mask int16
	for i := 0; i < paramsN/8; i++ {
		for j := 0; j < 8; j++ {
			mask = -int16((msg[i] >> j) & 1)
			r.coeffs[8*i+j] = mask & int16((paramsQ+1)/2)
		}
	}
	return r
}

func polyToMsg(a poly) []byte {
	msg := make([]byte, paramsSymBytes)
	var t uint16
	a = polyCSubQ(a)
	for i := 0; i < paramsN/8; i++ {
		msg[i] = 0
		for j := 0; j < 8; j++ {
			t = (((uint16(a.coeffs[8*i+j]) << 1) + uint16(paramsQ/2)) / uint16(paramsQ)) & 1
			msg[i] |= byte(t << j)
		}
	}
	return msg
}

func polyGetNoise(seed []byte, nonce byte) poly {
	l := paramsETA * paramsN / 4
	p := indcpaPrf(l, seed, nonce)
	return byteopsCbd(p)
}

func polyNtt(a poly) poly {
	r := polyNew()
	r.coeffs = ntt(a.coeffs)
	return polyReduce(r)
}

func polyInvNttToMont(a poly) poly {
	r := polyNew()
	r.coeffs = nttInv(a.coeffs)
	return r
}

func polyBaseMulMontgomery(a poly, b poly) poly {
	r := polyNew()
	for i := 0; i < paramsN/4; i++ {
		rx := nttBaseMul(
			a.coeffs[4*i+0], a.coeffs[4*i+1],
			b.coeffs[4*i+0], b.coeffs[4*i+1],
			nttZetas[64+i],
		)
		ry := nttBaseMul(
			a.coeffs[4*i+2], a.coeffs[4*i+3],
			b.coeffs[4*i+2], b.coeffs[4*i+3],
			-nttZetas[64+i],
		)
		r.coeffs[4*i+0] = rx[0]
		r.coeffs[4*i+1] = rx[1]
		r.coeffs[4*i+2] = ry[0]
		r.coeffs[4*i+3] = ry[1]
	}
	return r
}

func polyToMont(a poly) poly {
	r := polyNew()
	var f int16 = int16((uint64(1) << 32) % uint64(paramsQ))
	for i := 0; i < paramsN; i++ {
		r.coeffs[i] = byteopsMontgomeryReduce(int32(a.coeffs[i]) * int32(f))
	}
	return r
}

func polyReduce(a poly) poly {
	r := polyNew()
	for i := 0; i < paramsN; i++ {
		r.coeffs[i] = byteopsBarrettReduce(a.coeffs[i])
	}
	return r
}

func polyCSubQ(a poly) poly {
	r := polyNew()
	for i := 0; i < paramsN; i++ {
		r.coeffs[i] = byteopsCSubQ(a.coeffs[i])
	}
	return r
}

func polyAdd(a poly, b poly) poly {
	r := polyNew()
	for i := 0; i < paramsN; i++ {
		r.coeffs[i] = a.coeffs[i] + b.coeffs[i]
	}
	return r
}

func polySub(a poly, b poly) poly {
	r := polyNew()
	for i := 0; i < paramsN; i++ {
		r.coeffs[i] = a.coeffs[i] - b.coeffs[i]
	}
	return r
}

func polyvecCompress(a polyvec, paramsK int) []byte {
	var r []byte
	a = polyvecCSubQ(a, paramsK)
	rr := 0
	switch paramsK {
	case 2:
		r = make([]byte, paramsPolyvecCompressedBytesK2)
	case 3:
		r = make([]byte, paramsPolyvecCompressedBytesK3)
	case 4:
		r = make([]byte, paramsPolyvecCompressedBytesK4)
	}
	switch paramsK {
	case 2, 3:
		t := make([]uint16, 4)
		for i := 0; i < paramsK; i++ {
			for j := 0; j < paramsN/4; j++ {
				for k := 0; k < 4; k++ {
					t[k] = uint16((((uint32(a.vec[i].coeffs[4*j+k]) << 10) + uint32(paramsQ/2)) / uint32(paramsQ)) & 0x3ff)
				}
				r[rr+0] = byte(t[0] >> 0)
				r[rr+1] = byte((t[0] >> 8) | (t[1] << 2))
				r[rr+2] = byte((t[1] >> 6) | (t[2] << 4))
				r[rr+3] = byte((t[2] >> 4) | (t[3] << 6))
				r[rr+4] = byte((t[3] >> 2))
				rr = rr + 5
			}
		}
		return r
	default:
		t := make([]uint16, 8)
		for i := 0; i < paramsK; i++ {
			for j := 0; j < paramsN/8; j++ {
				for k := 0; k < 8; k++ {
					t[k] = uint16((((uint32(a.vec[i].coeffs[8*j+k]) << 11) + uint32(paramsQ/2)) / uint32(paramsQ)) & 0x7ff)
				}
				r[rr+0] = byte((t[0] >> 0))
				r[rr+1] = byte((t[0] >> 8) | (t[1] << 3))
				r[rr+2] = byte((t[1] >> 5) | (t[2] << 6))
				r[rr+3] = byte((t[2] >> 2))
				r[rr+4] = byte((t[2] >> 10) | (t[3] << 1))
				r[rr+5] = byte((t[3] >> 7) | (t[4] << 4))
				r[rr+6] = byte((t[4] >> 4) | (t[5] << 7))
				r[rr+7] = byte((t[5] >> 1))
				r[rr+8] = byte((t[5] >> 9) | (t[6] << 2))
				r[rr+9] = byte((t[6] >> 6) | (t[7] << 5))
				r[rr+10] = byte((t[7] >> 3))
				rr = rr + 11
			}
		}
		return r
	}
}

func polyvecDecompress(a []byte, paramsK int) polyvec {
	r := polyvecNew(paramsK)
	aa := 0
	switch paramsK {
	case 2, 3:
		t := make([]uint16, 4)
		for i := 0; i < paramsK; i++ {
			for j := 0; j < paramsN/4; j++ {
				t[0] = (uint16(a[aa+0]) >> 0) | (uint16(a[aa+1]) << 8)
				t[1] = (uint16(a[aa+1]) >> 2) | (uint16(a[aa+2]) << 6)
				t[2] = (uint16(a[aa+2]) >> 4) | (uint16(a[aa+3]) << 4)
				t[3] = (uint16(a[aa+3]) >> 6) | (uint16(a[aa+4]) << 2)
				aa = aa + 5
				for k := 0; k < 4; k++ {
					r.vec[i].coeffs[4*j+k] = int16((uint32(t[k]&0x3FF)*uint32(paramsQ) + 512) >> 10)
				}
			}
		}
	case 4:
		t := make([]uint16, 8)
		for i := 0; i < paramsK; i++ {
			for j := 0; j < paramsN/8; j++ {
				t[0] = (uint16(a[aa+0]) >> 0) | (uint16(a[aa+1]) << 8)
				t[1] = (uint16(a[aa+1]) >> 3) | (uint16(a[aa+2]) << 5)
				t[2] = (uint16(a[aa+2]) >> 6) | (uint16(a[aa+3]) << 2) | (uint16(a[aa+4]) << 10)
				t[3] = (uint16(a[aa+4]) >> 1) | (uint16(a[aa+5]) << 7)
				t[4] = (uint16(a[aa+5]) >> 4) | (uint16(a[aa+6]) << 4)
				t[5] = (uint16(a[aa+6]) >> 7) | (uint16(a[aa+7]) << 1) | (uint16(a[aa+8]) << 9)
				t[6] = (uint16(a[aa+8]) >> 2) | (uint16(a[aa+9]) << 6)
				t[7] = (uint16(a[aa+9]) >> 5) | (uint16(a[aa+10]) << 3)
				aa = aa + 11
				for k := 0; k < 8; k++ {
					r.vec[i].coeffs[8*j+k] = int16((uint32(t[k]&0x7FF)*uint32(paramsQ) + 1024) >> 11)
				}
			}
		}
	}
	return r
}

func polyvecToBytes(a polyvec, paramsK int) []byte {
	r := []byte{}
	for i := 0; i < paramsK; i++ {
		r = append(r, polyToBytes(a.vec[i])...)
	}
	return r
}

func polyvecFromBytes(a []byte, paramsK int) polyvec {
	r := polyvecNew(paramsK)
	for i := 0; i < paramsK; i++ {
		start := (i * paramsPolyBytes)
		end := (i + 1) * paramsPolyBytes
		r.vec[i] = polyFromBytes(a[start:end])
	}
	return r
}

func polyvecNtt(a polyvec, paramsK int) polyvec {
	r := polyvecNew(paramsK)
	for i := 0; i < paramsK; i++ {
		r.vec[i] = polyNtt(a.vec[i])
	}
	return r
}

func polyvecInvNttToMont(a polyvec, paramsK int) polyvec {
	r := polyvecNew(paramsK)
	for i := 0; i < paramsK; i++ {
		r.vec[i] = polyInvNttToMont(a.vec[i])
	}
	return r
}

func polyvecPointWiseAccMontgomery(a polyvec, b polyvec, paramsK int) poly {
	r := polyBaseMulMontgomery(a.vec[0], b.vec[0])
	for i := 1; i < paramsK; i++ {
		t := polyBaseMulMontgomery(a.vec[i], b.vec[i])
		r = polyAdd(r, t)
	}
	return polyReduce(r)
}

func polyvecReduce(a polyvec, paramsK int) polyvec {
	r := polyvecNew(paramsK)
	for i := 0; i < paramsK; i++ {
		r.vec[i] = polyReduce(a.vec[i])
	}
	return r
}

func polyvecCSubQ(a polyvec, paramsK int) polyvec {
	r := polyvecNew(paramsK)
	for i := 0; i < paramsK; i++ {
		r.vec[i] = polyCSubQ(a.vec[i])
	}
	return r
}

func polyvecAdd(a polyvec, b polyvec, paramsK int) polyvec {
	r := polyvecNew(paramsK)
	for i := 0; i < paramsK; i++ {
		r.vec[i] = polyAdd(a.vec[i], b.vec[i])
	}
	return r
}
