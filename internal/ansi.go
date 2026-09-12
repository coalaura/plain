package internal

const AnsiReset = "\x1b[0m"

func StripANSI(buf []byte) []byte {
	var j int

	for i := 0; i < len(buf); {
		// ESC sequence?
		if buf[i] == '\x1b' && i+1 < len(buf) {
			switch buf[i+1] {
			case '[':
				// CSI: ESC [ ... final_byte (0x40-0x7E)
				i += 2

				for i < len(buf) && (buf[i] < 0x40 || buf[i] > 0x7E) {
					i++
				}

				if i < len(buf) {
					i++ // skip the final byte too
				}

				continue
			case ']':
				// OSC: ESC ] ... BEL (0x07) or ST (ESC \)
				i += 2

				for i < len(buf) {
					if buf[i] == '\x07' {
						i++

						break
					}

					if buf[i] == '\x1b' && i+1 < len(buf) && buf[i+1] == '\\' {
						i += 2

						break
					}

					i++
				}

				continue
			default:
				// Other 2-byte escapes (ESC ( etc.)
				i += 2

				continue
			}
		}

		buf[j] = buf[i]

		j++
		i++
	}

	return buf[:j]
}
