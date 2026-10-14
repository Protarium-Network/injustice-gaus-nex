Scripts used to recover the constants in ../RECON.md from the retail
executable (`DCFGame-Retail.rpx`, not included). Run with `python -I`:

    python -I unrpx.py DCFGame-Retail.rpx dcf.bin    # zlib sections -> flat dump
    python -I findgid.py DCFGame-Retail.rpx          # section table -> secs.pkl
    python -I keyrefs.py                          # access-key literal -> code xref
    python -I dis.py 0x23120ec 0x2312150          # capstone PPC disassembly
    python -I refs.py "string"                    # ASCII string -> code xrefs

`dis.py` needs `capstone`. Outputs (`*.bin`, `secs.pkl`) are git-ignored.
