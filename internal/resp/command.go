package resp

type Command struct {
	Name string // GET, SET, AEGIS.*
	Key  string
	Args []string
	Raw  []byte // original RESP (important)
}

/*
parsing happens like:
cmd: SET user:123 value EX 300 ATAG users
Command{
    Name: "SET",
    Key:  "user:123",
    Args: ["value", "EX", "300", "ATAG", "users"],
    Raw:  [...],
}
*/
