package tianwen

// Embed the IANA database so timezone rules (including historical DST) do not
// depend on the host OS.
import _ "time/tzdata"
