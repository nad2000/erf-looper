// loop through ERF records wiht shifting ERF record timestamp
// in the first pass collects stats and calculates offset for the following passes

package multiplier

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//var (
//	Scale float
//)

func assert(t bool, s string) {
	if !t {
		panic(s)
	}
}

func ts_sec(ts int64) int64 {
	ts >>= 31
	return ts>>1 + (ts & 0x0000000000000001)
}

func Loop(in_f *os.File, out_f *os.File, off int, dry_run bool, limit int, count int) {

	var (
		global_offset                                       int64
		average_gap, gap_sum, offset                        int64
		ts, prev_ts, first_ts, last_ts, range_len, limit_ts int64
		wlen_sum                                            uint32
		rcount, icount                                      int
		// ERF record
		rec [100 * 1024]byte
	)

	if off != 0 {
		global_offset = int64(off) << 32
		fmt.Fprintf(os.Stderr, "** Offset: %d\n", ts_sec(global_offset))
	}

	incoming := make(chan os.Signal)
	signal.Notify(incoming, syscall.SIGTERM, syscall.SIGQUIT)

	first_pass := true

MAIN_LOOP:
	for {
		in_f.Seek(0, 0)
		icount++
		for {

			select {
			case sig := <-incoming:
				switch sig {
				// Process cancelled:
				case syscall.SIGTERM, syscall.SIGQUIT:
					break MAIN_LOOP
				}
			default:
				// Do nothing
			}

			_, err := in_f.Read(rec[:16])
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintln(os.Stderr, err)
				return
			}
			ts = int64(binary.LittleEndian.Uint64(rec[:8]))
			rlen := binary.BigEndian.Uint16(rec[10:12])
			//fmt.Fprintf(os.Stderr, "rlen: %d\n", rlen)

			wlen := binary.BigEndian.Uint16(rec[14:16])
			if first_pass {
				rcount++
				wlen_sum += uint32(wlen)
				if first_ts == 0 {
					first_ts = ts
					if limit != 0 {
						limit_ts = first_ts + int64(limit)<<32
					}
				}
			}
			// fmt.Fprintf(os.Stderr, "Read ts %d length %d\n", uint32(ts >> 32), rlen)
			// if rcount % 100 == 0 {
			//     os.Stderr.WriteString("*")
			//     //fmt.Fprintf(os.Stderr, "Read ts %d length %d\n", uint32(ts >> 32), rlen)
			// }
			bcount, err := in_f.Read(rec[16:rlen])
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintln(os.Stderr, err)
				return
			}
			// replace rlen with actually read byte count
			if bcount+16 < int(rlen) {
				rlen = uint16(bcount + 16)
			}

			new_ts := ts + offset + global_offset
			if new_ts < prev_ts {
				fmt.Fprintf(os.Stderr,
					"WARNING! Unexpected record with timestamp %d (%d) - goes back in time. It should exceed %d (%d)\n",
					new_ts, new_ts<<32, prev_ts, prev_ts<<32)
			}

			binary.LittleEndian.PutUint64(rec[:8], uint64(new_ts))

			if !dry_run {
				if limit_ts != 0 && ts+offset > limit_ts {
					return
				}
				out_f.Write(rec[:rlen])
			}

			if first_pass {
				if last_ts > 0 {
					gap_sum += (ts - last_ts)
				}
				last_ts = ts
			}
		}

		// Calucate metrics only on the first pass
		if first_pass {
			range_len = last_ts - first_ts
			average_gap = gap_sum / int64(rcount-1)

			fmt.Fprintf(os.Stderr,
				"** Total %d sec (%d packets), %d bytes, average %d bytes per packet\n** from %v to %v (from: %d, to: %d)\n",
				ts_sec(range_len), rcount, wlen_sum, wlen_sum/uint32(rcount),
				time.Unix(ts_sec(first_ts), 0).UTC(), time.Unix(ts_sec(last_ts), 0).UTC(),
				ts_sec(first_ts), ts_sec(last_ts))

			if global_offset != 0 {
				fmt.Fprintf(os.Stderr,
					"** Modified time range: from %v to %v (from: %d, to: %d)\n",
					time.Unix(ts_sec(first_ts+global_offset), 0).UTC(), time.Unix(ts_sec(last_ts+global_offset), 0).UTC(),
					ts_sec(first_ts+global_offset), ts_sec(last_ts+global_offset))
			}
			if dry_run {
				break MAIN_LOOP
			}

		}
		offset += average_gap + range_len
		os.Stderr.WriteString("*")
		// fmt.Fprintf( os.Stderr, "** offset: %d sec, gap: %d\n", offset >> 32, average_gap)

		first_pass = false

		if (count != 0 && icount >= count) {
			break MAIN_LOOP
		}

	}

}
