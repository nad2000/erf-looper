package multiplier

import (
    "testing"
    "encoding/binary"
    "os"
    "bytes"
)

var test_recs = [][]byte {
    //   time stamp    |type|flags|rlen | lctr|wlen| ... (the rest)
    // 0 1 2 3  0 1 2 3
    {  0,1,0,0, 0,0,0,0, 1,   2,    0,20, 0,17, 7,8, 0xFF, 0xF0, 0xF1, 0xF2 },
    {  0,1,0,0, 1,0,0,0, 1,   2,    0,19, 0,17, 7,8, 0xF3, 0xF4, 0x6F },
    {  0,1,0,0, 2,0,0,0, 1,   2,    0,20, 0,17, 7,8, 0xEF, 0xFF, 0xFF, 0xFF },
    {  0,1,0,0, 3,0,0,0, 1,   2,    0,18, 0,17, 7,8, 0x1F, 0xFF },
}

func ts( rec []byte ) int64 {
    return int64(binary.LittleEndian.Uint64(rec[:8]) )
}

func TestERFLoop(t *testing.T) {

    offset := 123
    var avg_gap int64

    for i := 0; i < len(test_recs)-1; i++ {
        avg_gap += ts(test_recs[i+1])-ts(test_recs[i])
    }
    avg_gap /= int64(len(test_recs)-1)

    span := ts(test_recs[len(test_recs)-1])-ts(test_recs[0])

    t.Logf("AVG Gap: %d (%d), span: %d (%d), offset: %d", avg_gap, avg_gap>>32, span, span>>32, offset)


    in_r, in_w, _ := os.Pipe()
    out_r, out_w, _ := os.Pipe()
    go Loop( in_r, out_w, offset, 0)

    // read the first pass output
    rec := make([]byte,100)
    for _, r := range test_recs {
        in_w.Write(r)
        out_r.Read(rec[:len(r)])
        old_ts, new_ts := ts(r), ts(rec)
        // timestamp should be shifted
        expected_ts := old_ts + int64(offset) << 32
        if expected_ts != new_ts {
            t.Errorf("new timestamp: %d, want %d (%d sec, %d sec)\n** Original packed:\t%v\n** Modified:\t\t%v",
                new_ts, old_ts, new_ts>>32, old_ts>>32, r, rec[:len(r)] )
        }
        // the rest should be identical
        if !bytes.Equal( r[8:], rec[8:len(r)]) {
            t.Errorf("The rest of the r shold not be modified (except timestamp).\n** Original packed:\t%v\n** Modified:\t\t%v",
                r[8:], rec[8:len(r)] )
        }
    }

    // // TODO: second pass
    // for _, r := range test_recs {
    //     in_w.Write(r)
    //     out_r.Read(rec[:len(r)])
    //     old_ts, new_ts := ts(r), ts(rec)
    //     expected_ts := old_ts + int64(offset) << 32 + avg_gap+span
    //     // timestamp should be shifted (offset + average gap + span)
    //     if expected_ts != new_ts {
    //         t.Errorf("new timestamp: %d, want %d (%d sec, %d sec)\n** Original packed:\t%v\n** Modified:\t\t%v",
    //             new_ts, expected_ts, new_ts>>32, expected_ts>>32, r, rec[:len(r)] )
    //     }
    //     // the rest should be identical
    //     if !bytes.Equal( r[8:], rec[8:len(r)]) {
    //         t.Errorf("The rest of the r shold not be modified (except timestamp).\n** Original packed:\t%v\n** Modified:\t\t%v",
    //             r[8:], rec[8:len (r)] )
    //     }
    // }

}