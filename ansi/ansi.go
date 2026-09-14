package ansi

const (
	// AnsiReset resets all terminal attributes and colors.
	AnsiReset = "\x1b[0m"

	// AnsiBlack sets the foreground color to normal-intensity black.
	AnsiBlack = "\x1b[30m"

	// AnsiRed sets the foreground color to normal-intensity red.
	AnsiRed = "\x1b[31m"

	// AnsiGreen sets the foreground color to normal-intensity green.
	AnsiGreen = "\x1b[32m"

	// AnsiYellow sets the foreground color to normal-intensity yellow.
	AnsiYellow = "\x1b[33m"

	// AnsiBlue sets the foreground color to normal-intensity blue.
	AnsiBlue = "\x1b[34m"

	// AnsiMagenta sets the foreground color to normal-intensity magenta.
	AnsiMagenta = "\x1b[35m"

	// AnsiCyan sets the foreground color to normal-intensity cyan.
	AnsiCyan = "\x1b[36m"

	// AnsiWhite sets the foreground color to normal-intensity white.
	AnsiWhite = "\x1b[37m"

	// AnsiHiBlack sets the foreground color to high-intensity black.
	// This is usually rendered as gray or dark gray.
	AnsiHiBlack = "\x1b[90m"

	// AnsiHiRed sets the foreground color to high-intensity red.
	AnsiHiRed = "\x1b[91m"

	// AnsiHiGreen sets the foreground color to high-intensity green.
	AnsiHiGreen = "\x1b[92m"

	// AnsiHiYellow sets the foreground color to high-intensity yellow.
	AnsiHiYellow = "\x1b[93m"

	// AnsiHiBlue sets the foreground color to high-intensity blue.
	AnsiHiBlue = "\x1b[94m"

	// AnsiHiMagenta sets the foreground color to high-intensity magenta.
	AnsiHiMagenta = "\x1b[95m"

	// AnsiHiCyan sets the foreground color to high-intensity cyan.
	AnsiHiCyan = "\x1b[96m"

	// AnsiHiWhite sets the foreground color to high-intensity white.
	AnsiHiWhite = "\x1b[97m"
)

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
