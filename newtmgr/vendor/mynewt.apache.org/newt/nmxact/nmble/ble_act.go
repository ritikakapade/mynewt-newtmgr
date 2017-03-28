package nmble

import (
	"encoding/json"

	"mynewt.apache.org/newt/nmxact/nmxutil"
)

// Blocking
func connect(x *BleXport, connChan chan error, r *BleConnectReq) error {
	j, err := json.Marshal(r)
	if err != nil {
		return err
	}

	if err := x.Tx(j); err != nil {
		return err
	}

	err = <-connChan
	if err != nil {
		return err
	}

	return nil
}

// Blocking
func terminate(x *BleXport, bl *BleListener, r *BleTerminateReq) error {
	j, err := json.Marshal(r)
	if err != nil {
		return err
	}

	if err := x.Tx(j); err != nil {
		return err
	}

	for {
		select {
		case err := <-bl.ErrChan:
			return err

		case bm := <-bl.BleChan:
			switch msg := bm.(type) {
			case *BleTerminateRsp:
				bl.Acked = true
				if msg.Status != 0 {
					return StatusError(MSG_OP_RSP,
						MSG_TYPE_TERMINATE,
						msg.Status)
				} else {
					return nil
				}

			default:
			}

		case <-bl.AfterTimeout(x.rspTimeout):
			return BhdTimeoutError(MSG_TYPE_TERMINATE)
		}
	}
}

func connCancel(x *BleXport, bl *BleListener, r *BleConnCancelReq) error {
	j, err := json.Marshal(r)
	if err != nil {
		return err
	}

	if err := x.Tx(j); err != nil {
		return err
	}

	for {
		select {
		case err := <-bl.ErrChan:
			return err

		case bm := <-bl.BleChan:
			switch msg := bm.(type) {
			case *BleConnCancelRsp:
				bl.Acked = true
				if msg.Status != 0 {
					return StatusError(MSG_OP_RSP,
						MSG_TYPE_CONN_CANCEL,
						msg.Status)
				} else {
					return nil
				}

			default:
			}

		case <-bl.AfterTimeout(x.rspTimeout):
			return BhdTimeoutError(MSG_TYPE_TERMINATE)
		}
	}
}

// Blocking.
func discSvcUuid(x *BleXport, bl *BleListener, r *BleDiscSvcUuidReq) (
	*BleSvc, error) {

	j, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	if err := x.Tx(j); err != nil {
		return nil, err
	}

	var svc *BleSvc
	for {
		select {
		case err := <-bl.ErrChan:
			return nil, err

		case bm := <-bl.BleChan:
			switch msg := bm.(type) {
			case *BleDiscSvcUuidRsp:
				bl.Acked = true
				if msg.Status != 0 {
					return nil, StatusError(MSG_OP_RSP,
						MSG_TYPE_DISC_SVC_UUID,
						msg.Status)
				}

			case *BleDiscSvcEvt:
				switch msg.Status {
				case 0:
					svc = &msg.Svc
				case ERR_CODE_EDONE:
					if svc == nil {
						return nil, nmxutil.FmtBleHostError(
							msg.Status,
							"Peer doesn't support required service: %s",
							r.Uuid.String())
					}
					return svc, nil
				default:
					return nil, StatusError(MSG_OP_EVT,
						MSG_TYPE_DISC_SVC_EVT,
						msg.Status)
				}

			default:
			}

		case <-bl.AfterTimeout(x.rspTimeout):
			return nil, BhdTimeoutError(MSG_TYPE_DISC_SVC_UUID)
		}
	}
}

// Blocking.
func discAllChrs(x *BleXport, bl *BleListener, r *BleDiscAllChrsReq) (
	[]*BleChr, error) {

	j, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	if err := x.Tx(j); err != nil {
		return nil, err
	}

	chrs := []*BleChr{}
	for {
		select {
		case err := <-bl.ErrChan:
			return nil, err

		case bm := <-bl.BleChan:
			switch msg := bm.(type) {
			case *BleDiscAllChrsRsp:
				bl.Acked = true
				if msg.Status != 0 {
					return nil, StatusError(MSG_OP_RSP,
						MSG_TYPE_DISC_ALL_CHRS,
						msg.Status)
				}

			case *BleDiscChrEvt:
				switch msg.Status {
				case 0:
					chrs = append(chrs, &msg.Chr)
				case ERR_CODE_EDONE:
					return chrs, nil
				default:
					return nil, StatusError(MSG_OP_EVT,
						MSG_TYPE_DISC_CHR_EVT,
						msg.Status)
				}

			default:
			}

		case <-bl.AfterTimeout(x.rspTimeout):
			return nil, BhdTimeoutError(MSG_TYPE_DISC_ALL_CHRS)
		}
	}
}

// Blocking.
func writeCmd(x *BleXport, bl *BleListener, r *BleWriteCmdReq) error {
	j, err := json.Marshal(r)
	if err != nil {
		return err
	}

	if err := x.Tx(j); err != nil {
		return err
	}

	for {
		select {
		case err := <-bl.ErrChan:
			return err

		case bm := <-bl.BleChan:
			switch msg := bm.(type) {
			case *BleWriteCmdRsp:
				bl.Acked = true
				if msg.Status != 0 {
					return StatusError(MSG_OP_RSP,
						MSG_TYPE_WRITE_CMD,
						msg.Status)
				} else {
					return nil
				}

			default:
			}

		case <-bl.AfterTimeout(x.rspTimeout):
			return BhdTimeoutError(MSG_TYPE_WRITE_CMD)
		}
	}
}

// Blocking.
func exchangeMtu(x *BleXport, bl *BleListener, r *BleExchangeMtuReq) (
	int, error) {

	j, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	if err := x.Tx(j); err != nil {
		return 0, err
	}

	for {
		select {
		case err := <-bl.ErrChan:
			return 0, err

		case bm := <-bl.BleChan:
			switch msg := bm.(type) {
			case *BleExchangeMtuRsp:
				bl.Acked = true
				if msg.Status != 0 {
					return 0, StatusError(MSG_OP_RSP,
						MSG_TYPE_EXCHANGE_MTU,
						msg.Status)
				}

			case *BleMtuChangeEvt:
				if msg.Status != 0 {
					return 0, StatusError(MSG_OP_EVT,
						MSG_TYPE_MTU_CHANGE_EVT,
						msg.Status)
				} else {
					return msg.Mtu, nil
				}

			default:
			}

		case <-bl.AfterTimeout(x.rspTimeout):
			return 0, BhdTimeoutError(MSG_TYPE_EXCHANGE_MTU)
		}
	}
}

type scanFn func(evt *BleScanEvt)

func scan(x *BleXport, bl *BleListener, r *BleScanReq,
	abortChan chan struct{}, scanCb scanFn) error {

	j, err := json.Marshal(r)
	if err != nil {
		return err
	}

	if err := x.Tx(j); err != nil {
		return err
	}

	for {
		select {
		case err := <-bl.ErrChan:
			return err

		case bm := <-bl.BleChan:
			switch msg := bm.(type) {
			case *BleScanRsp:
				bl.Acked = true
				if msg.Status != 0 {
					return StatusError(MSG_OP_RSP, MSG_TYPE_SCAN, msg.Status)
				}

			case *BleScanEvt:
				scanCb(msg)

			default:
			}

		case <-bl.AfterTimeout(x.rspTimeout):
			return BhdTimeoutError(MSG_TYPE_EXCHANGE_MTU)

		case <-abortChan:
			return nil
		}
	}
}

func scanCancel(x *BleXport, bl *BleListener, r *BleScanCancelReq) error {
	j, err := json.Marshal(r)
	if err != nil {
		return err
	}

	if err := x.Tx(j); err != nil {
		return err
	}

	for {
		select {
		case err := <-bl.ErrChan:
			return err

		case bm := <-bl.BleChan:
			switch msg := bm.(type) {
			case *BleScanCancelRsp:
				bl.Acked = true
				if msg.Status != 0 {
					return StatusError(MSG_OP_RSP, MSG_TYPE_SCAN, msg.Status)
				}
				return nil

			default:
			}

		case <-bl.AfterTimeout(x.rspTimeout):
			return BhdTimeoutError(MSG_TYPE_EXCHANGE_MTU)
		}
	}
}