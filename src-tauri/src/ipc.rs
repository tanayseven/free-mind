//! Messages between the app and the daemon: one JSON object per line over a
//! local socket (Unix domain socket, or a named pipe on Windows).

use interprocess::local_socket::{prelude::*, Name, Stream};
use serde::{Deserialize, Serialize};
use std::io::{self, BufRead, BufReader, Read, Write};

#[cfg(unix)]
pub const SOCKET_PATH: &str = if cfg!(target_os = "macos") {
    "/var/run/tech.tanay.free-mind.sock"
} else {
    "/run/tech.tanay.free-mind.sock"
};

#[cfg(windows)]
const PIPE_NAME: &str = "tech.tanay.free-mind";

/// Upper bound on a single message, so a client can't exhaust daemon memory.
pub const MAX_MESSAGE_BYTES: u64 = 1024 * 1024;

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(tag = "cmd", rename_all = "lowercase")]
pub enum Request {
    Ping,
    Block { domains: Vec<String> },
    Unblock,
}

#[derive(Serialize, Deserialize, Debug, PartialEq)]
pub struct Response {
    pub ok: bool,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

impl Response {
    pub fn ok() -> Self {
        Response {
            ok: true,
            error: None,
        }
    }

    pub fn err(message: impl Into<String>) -> Self {
        Response {
            ok: false,
            error: Some(message.into()),
        }
    }
}

pub fn socket_name() -> io::Result<Name<'static>> {
    #[cfg(unix)]
    {
        use interprocess::local_socket::GenericFilePath;
        SOCKET_PATH.to_fs_name::<GenericFilePath>()
    }
    #[cfg(windows)]
    {
        use interprocess::local_socket::GenericNamespaced;
        PIPE_NAME.to_ns_name::<GenericNamespaced>()
    }
}

/// Reads one newline-terminated JSON message.
pub fn read_message<T: for<'de> Deserialize<'de>>(reader: &mut impl BufRead) -> io::Result<T> {
    let mut line = String::new();
    reader.take(MAX_MESSAGE_BYTES).read_line(&mut line)?;
    if line.is_empty() {
        return Err(io::ErrorKind::UnexpectedEof.into());
    }
    serde_json::from_str(&line).map_err(io::Error::other)
}

/// Writes one newline-terminated JSON message.
pub fn write_message<T: Serialize>(writer: &mut impl Write, message: &T) -> io::Result<()> {
    let mut line = serde_json::to_vec(message).map_err(io::Error::other)?;
    line.push(b'\n');
    writer.write_all(&line)?;
    writer.flush()
}

/// Sends `request` to the daemon and waits for its reply.
pub fn send(request: &Request) -> Result<(), String> {
    let stream = Stream::connect(socket_name().map_err(|e| e.to_string())?)
        .map_err(|e| format!("could not reach the Free Mind daemon: {e}"))?;
    let mut stream = BufReader::new(stream);
    write_message(stream.get_mut(), request).map_err(|e| e.to_string())?;
    let response: Response = read_message(&mut stream).map_err(|e| e.to_string())?;
    match response {
        Response { ok: true, .. } => Ok(()),
        Response { error, .. } => Err(error.unwrap_or_else(|| "daemon request failed".into())),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn request_wire_format() {
        let request = Request::Block {
            domains: vec!["x.com".into()],
        };
        let json = serde_json::to_string(&request).unwrap();
        assert_eq!(json, r#"{"cmd":"block","domains":["x.com"]}"#);
        assert_eq!(serde_json::from_str::<Request>(&json).unwrap(), request);
        assert_eq!(
            serde_json::from_str::<Request>(r#"{"cmd":"ping"}"#).unwrap(),
            Request::Ping
        );
    }

    #[test]
    fn response_round_trip() {
        assert_eq!(
            serde_json::to_string(&Response::ok()).unwrap(),
            r#"{"ok":true}"#
        );
        let mut buf = Vec::new();
        write_message(&mut buf, &Response::err("nope")).unwrap();
        let read: Response = read_message(&mut buf.as_slice()).unwrap();
        assert_eq!(read, Response::err("nope"));
    }

    #[test]
    fn read_message_rejects_eof_and_garbage() {
        assert!(read_message::<Request>(&mut &b""[..]).is_err());
        assert!(read_message::<Request>(&mut &b"{\"cmd\":\"rm -rf\"}\n"[..]).is_err());
    }
}
