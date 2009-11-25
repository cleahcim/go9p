// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import "syscall"

// Creates a Fcall value from the on-the-wire representation. If
// dotu is true, reads 9P2000.u messages. Returns the unpacked message,
// error and how many bytes from the buffer were used by the message.
func Unpack(buf []byte, dotu bool) (fc *Fcall, err *Error, fcsz int) {
	var m uint16;

	fc = new(Fcall);
	fc.Fid = Nofid;
	fc.Afid = Nofid;
	fc.Newfid = Nofid;

	p := buf;
	fc.size, p = gint32(p);
	fc.Id, p = gint8(p);
	fc.Tag, p = gint16(p);

	p = p[0 : fc.size-7];
	fc.Pkt = buf[0:fc.size];
	fcsz = int(fc.size);
	if fc.Id < Tfirst || fc.Id >= Tlast {
		return nil, &Error{"invalid id", syscall.EINVAL}, 0
	}

	var sz uint32;
	if dotu {
		sz = minFcsize[fc.Id-Tfirst]
	} else {
		sz = minFcusize[fc.Id-Tfirst]
	}

	if fc.size < sz {
	szerror:
		return nil, &Error{"invalid size", syscall.EINVAL}, 0
	}

	err = nil;
	switch fc.Id {
	default:
		return nil, &Error{"invalid message id", syscall.EINVAL}, 0

	case Tversion, Rversion:
		fc.Msize, p = gint32(p);
		fc.Version, p = gstr(p);
		if p == nil {
			goto szerror
		}

	case Tauth:
		fc.Afid, p = gint32(p);
		fc.Uname, p = gstr(p);
		if p == nil {
			goto szerror
		}

		fc.Aname, p = gstr(p);
		if p == nil {
			goto szerror
		}

		if dotu {
			if len(p) > 0 {
				fc.Nuname, p = gint32(p)
			} else {
				fc.Nuname = Nouid
			}
		} else {
			fc.Nuname = Nouid
		}

	case Rauth, Rattach:
		p = gqid(p, &fc.Fqid)

	case Tflush:
		fc.Oldtag, p = gint16(p)

	case Tattach:
		fc.Fid, p = gint32(p);
		fc.Afid, p = gint32(p);
		fc.Uname, p = gstr(p);
		if p == nil {
			goto szerror
		}

		fc.Aname, p = gstr(p);
		if p == nil {
			goto szerror
		}

		if dotu {
			if len(p) > 0 {
				fc.Nuname, p = gint32(p)
			} else {
				fc.Nuname = Nouid
			}
		}

	case Rerror:
		fc.Error, p = gstr(p);
		if p == nil {
			goto szerror
		}

		if dotu {
			fc.Nerror, p = gint32(p)
		} else {
			fc.Nerror = 0
		}

	case Twalk:
		fc.Fid, p = gint32(p);
		fc.Newfid, p = gint32(p);
		m, p = gint16(p);
		fc.Wnames = make([]string, m);
		for i := 0; i < int(m); i++ {
			fc.Wnames[i], p = gstr(p);
			if p == nil {
				goto szerror
			}
		}

	case Rwalk:
		m, p = gint16(p);
		fc.Wqids = make([]Qid, m);
		for i := 0; i < int(m); i++ {
			p = gqid(p, &fc.Wqids[i])
		}

	case Topen:
		fc.Fid, p = gint32(p);
		fc.Mode, p = gint8(p);

	case Ropen, Rcreate:
		p = gqid(p, &fc.Fqid);
		fc.Iounit, p = gint32(p);

	case Tcreate:
		fc.Fid, p = gint32(p);
		fc.Name, p = gstr(p);
		if p == nil {
			goto szerror
		}
		fc.Perm, p = gint32(p);
		fc.Mode, p = gint8(p);
		if dotu {
			fc.Ext, p = gstr(p);
			if p == nil {
				goto szerror
			}
		}

	case Tread:
		fc.Fid, p = gint32(p);
		fc.Offset, p = gint64(p);
		fc.Count, p = gint32(p);

	case Rread:
		fc.Count, p = gint32(p);
		if len(p) < int(fc.Count) {
			goto szerror
		}
		fc.Data = p;
		p = p[fc.Count:len(p)];

	case Twrite:
		fc.Fid, p = gint32(p);
		fc.Offset, p = gint64(p);
		fc.Count, p = gint32(p);
		if len(p) != int(fc.Count) {
			goto szerror
		}
		fc.Data = p;
		p = p[fc.Count:len(p)];

	case Rwrite:
		fc.Count, p = gint32(p)

	case Tclunk, Tremove, Tstat:
		fc.Fid, p = gint32(p)

	case Rstat:
		m, p = gint16(p);
		p = gstat(p, &fc.Fdir, dotu);
		if p == nil {
			goto szerror
		}

	case Twstat:
		fc.Fid, p = gint32(p);
		m, p = gint16(p);
		p = gstat(p, &fc.Fdir, dotu);

	case Rflush, Rclunk, Rremove, Rwstat:
	}

	if len(p) > 0 {
		goto szerror
	}

	return;
}