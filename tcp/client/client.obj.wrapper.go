package client

//type clientObjWrapper struct {
//	*clientObj
//}
//
//func (c *clientObjWrapper) Logger() log.Logger {
//	return c.clientObj.Logger
//}
//
//func (c *clientObjWrapper) Close() {
//	_ = c.clientObj.conn.Close()
//}
//
//func (c *clientObjWrapper) RawWrite(ctx context.Context, message []byte) (n int, err error) {
//	n, err = c.clientObj.conn.Write(message)
//	return
//}
