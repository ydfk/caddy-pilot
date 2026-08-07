import type { PasskeyCreationOptionsJSON, PasskeyRequestOptionsJSON } from "@/api/auth";

export function isPasskeySupported() {
  return window.isSecureContext && "PublicKeyCredential" in window && !!navigator.credentials;
}

export async function createPasskey(options: PasskeyCreationOptionsJSON) {
  ensurePasskeySupport();
  const credential = await navigator.credentials.create({
    publicKey: {
      ...options,
      challenge: decodeBase64URL(options.challenge),
      user: { ...options.user, id: decodeBase64URL(options.user.id) },
      excludeCredentials: options.excludeCredentials?.map((item) => ({
        ...item,
        id: decodeBase64URL(item.id),
      })),
    },
  });
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error("浏览器没有返回有效的 Passkey 凭据");
  }
  const response = credential.response;
  if (!(response instanceof AuthenticatorAttestationResponse)) {
    throw new Error("Passkey 注册响应类型无效");
  }
  return {
    id: credential.id,
    type: credential.type,
    rawId: encodeBase64URL(credential.rawId),
    authenticatorAttachment: credential.authenticatorAttachment,
    clientExtensionResults: credential.getClientExtensionResults(),
    response: {
      attestationObject: encodeBase64URL(response.attestationObject),
      clientDataJSON: encodeBase64URL(response.clientDataJSON),
      transports: response.getTransports?.() ?? [],
    },
  };
}

export async function getPasskey(options: PasskeyRequestOptionsJSON) {
  ensurePasskeySupport();
  const credential = await navigator.credentials.get({
    publicKey: {
      ...options,
      challenge: decodeBase64URL(options.challenge),
      allowCredentials: options.allowCredentials?.map((item) => ({
        ...item,
        id: decodeBase64URL(item.id),
      })),
    },
  });
  if (!(credential instanceof PublicKeyCredential)) {
    throw new Error("浏览器没有返回有效的 Passkey 凭据");
  }
  const response = credential.response;
  if (!(response instanceof AuthenticatorAssertionResponse)) {
    throw new Error("Passkey 登录响应类型无效");
  }
  return {
    id: credential.id,
    type: credential.type,
    rawId: encodeBase64URL(credential.rawId),
    authenticatorAttachment: credential.authenticatorAttachment,
    clientExtensionResults: credential.getClientExtensionResults(),
    response: {
      authenticatorData: encodeBase64URL(response.authenticatorData),
      clientDataJSON: encodeBase64URL(response.clientDataJSON),
      signature: encodeBase64URL(response.signature),
      userHandle: response.userHandle ? encodeBase64URL(response.userHandle) : null,
    },
  };
}

export function passkeyErrorMessage(reason: unknown) {
  if (reason instanceof DOMException && reason.name === "NotAllowedError") {
    return "Passkey 操作已取消或等待超时";
  }
  return reason instanceof Error ? reason.message : "Passkey 操作失败，请稍后重试";
}

function ensurePasskeySupport() {
  if (!window.isSecureContext) {
    throw new Error("Passkey 需要 HTTPS；只有 localhost 可以使用 HTTP");
  }
  if (!("PublicKeyCredential" in window) || !navigator.credentials) {
    throw new Error("当前浏览器不支持 Passkey");
  }
}

function decodeBase64URL(value: string) {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, "=");
  const binary = atob(padded);
  return Uint8Array.from(binary, (character) => character.charCodeAt(0));
}

function encodeBase64URL(value: ArrayBuffer) {
  const bytes = new Uint8Array(value);
  let binary = "";
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}
