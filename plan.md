# Tailscale + Custom Domain Plan

## Problem

Want `dev.argocd.synapse.mabbott.dev` accessible from any device on tailnet (phone, computer, etc).

## Current State

- Tailscale Ingress works from phone via `synapse-gateway-1.tail93695b.ts.net`
- Custom domain fails due to TLS cert mismatch (Tailscale cert is for `.ts.net` only)
- Service annotation approach doesn't work from phone over DERP relay

## Root Cause

Tailscale Ingress terminates TLS with a cert for `*.ts.net`. When accessing via custom domain, TLS handshake fails.

## Solution

kgateway should terminate TLS with a cert for the custom domain. Tailscale should do TCP passthrough, not TLS termination.

### Target Architecture

```
Phone → Tailscale (TCP passthrough) → kgateway (TLS termination with custom cert) → ArgoCD
```

### Implementation Steps

1. **Install cert-manager** on axon cluster (if not already present)

2. **Create ClusterIssuer** for Let's Encrypt (or self-signed for internal use)

3. **Create Certificate** for `dev.argocd.synapse.mabbott.dev`

4. **Update kgateway Gateway** to use HTTPS listener with the cert

5. **Configure Tailscale for TCP passthrough** instead of TLS termination
   - Option A: Use `tailscale.com/expose` annotation with TCP mode
   - Option B: Configure Tailscale Ingress for TCP passthrough

6. **Update DNS** - A record pointing `dev.argocd.synapse.mabbott.dev` to Tailscale IP

7. **Test** from phone and computer

## Open Questions

- Do we want Let's Encrypt (requires DNS challenge for internal domains) or self-signed cert?
- Can Tailscale Ingress do TCP passthrough, or do we need a different approach?
