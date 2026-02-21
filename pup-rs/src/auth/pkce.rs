use anyhow::Result;
use base64::{engine::general_purpose::URL_SAFE_NO_PAD, Engine};
use rand::RngCore;
use sha2::{Digest, Sha256};

pub struct PkceChallenge {
    pub verifier: String,
    pub challenge: String,
    pub method: String,
}

/// Generate a PKCE S256 challenge (RFC 7636).
pub fn generate_pkce_challenge() -> Result<PkceChallenge> {
    let verifier = generate_random_string(128)?;
    let challenge = {
        let hash = Sha256::digest(verifier.as_bytes());
        URL_SAFE_NO_PAD.encode(hash)
    };
    Ok(PkceChallenge {
        verifier,
        challenge,
        method: "S256".to_string(),
    })
}

/// Generate a random state parameter for CSRF protection.
pub fn generate_state() -> Result<String> {
    generate_random_string(32)
}

fn generate_random_string(length: usize) -> Result<String> {
    let byte_len = (length * 3) / 4 + 1;
    let mut bytes = vec![0u8; byte_len];
    rand::thread_rng().fill_bytes(&mut bytes);
    let encoded = URL_SAFE_NO_PAD.encode(&bytes);
    Ok(encoded[..length].to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_pkce_challenge() {
        let challenge = generate_pkce_challenge().unwrap();
        assert_eq!(challenge.verifier.len(), 128);
        assert!(!challenge.challenge.is_empty());
        assert_eq!(challenge.method, "S256");
        // Verify the challenge is the base64url-encoded SHA256 of the verifier
        let expected = {
            let hash = Sha256::digest(challenge.verifier.as_bytes());
            URL_SAFE_NO_PAD.encode(hash)
        };
        assert_eq!(challenge.challenge, expected);
    }

    #[test]
    fn test_state_length() {
        let state = generate_state().unwrap();
        assert_eq!(state.len(), 32);
    }

    #[test]
    fn test_randomness() {
        let a = generate_state().unwrap();
        let b = generate_state().unwrap();
        assert_ne!(a, b);
    }
}
