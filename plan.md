# Tailscale + Custom Domain Plan

## Problem

Want `dev.argocd.synapse.mabbott.dev` accessible from any device on tailnet (phone, computer, etc).

## Current State

- Tailscale Ingress works from phone via `synapse-gateway-1.tail93695b.ts.net`
- Custom domain fails due to TLS cert mismatch (Tailscale cert is for `.ts.net` only)
- Service annotation approach (`tailscale.com/expose`) doesn't work from phone over DERP relay

## Root Cause

Tailscale Ingress terminates TLS with a cert for `*.ts.net`. When accessing via custom domain, TLS handshake fails.

## Tailscale Operator Modes

| Mode | Resource | TLS Behavior | Works from phone? |
|------|----------|--------------|-------------------|
| Ingress | `Ingress` with `ingressClassName: tailscale` | Terminates TLS (`.ts.net` cert) | ✓ Yes |
| LoadBalancer | Service annotation `tailscale.com/expose: "true"` | TCP passthrough (no TLS) | ✗ No (DERP relay issue?) |

## Solution

kgateway should terminate TLS with a cert for the custom domain. Tailscale should do TCP passthrough.

**Problem:** The LoadBalancer/annotation approach (TCP passthrough) doesn't work from phone. Only the Ingress approach works, but it terminates TLS with wrong cert.

### Target Architecture

```
Phone → Tailscale (TCP passthrough) → kgateway (TLS termination with custom cert) → ArgoCD
```

### Implementation Options

#### Option A: Fix LoadBalancer approach (TCP passthrough)
- Figure out why `tailscale.com/expose` annotation doesn't work from phone over DERP relay
- Once working, kgateway handles TLS with custom domain cert
- Blocker: Unknown why it fails from phone

#### Option B: Use Tailscale Ingress with custom cert
- Configure Tailscale Ingress to use custom TLS cert instead of auto-provisioned `.ts.net` cert
- Need to check if this is supported
- Tailscale docs mention: "TLS certificates and renewal - The operator automatically provisions TLS certificates"
- May not be configurable

#### Option C: cert-manager + kgateway HTTPS + Tailscale Ingress passthrough
- Tailscale Ingress → kgateway port 443 (HTTPS)
- kgateway terminates TLS with custom cert from cert-manager
- Issue: Tailscale Ingress expects to terminate TLS itself

#### Option D: Split DNS with Pi-hole/CoreDNS
- Run DNS server on tailnet
- Resolve `dev.argocd.synapse.mabbott.dev` → Tailscale IP internally
- Access via `.ts.net` hostname, DNS just makes it pretty
- Requires running additional infrastructure

### Implementation Steps (Option C - most promising)

1. **Install cert-manager** on axon cluster

2. **Create ClusterIssuer** for Let's Encrypt with DNS01 challenge (Cloudflare)

3. **Create Certificate** for `*.dev.synapse.mabbott.dev`

4. **Update kgateway Gateway** to use HTTPS listener on port 443 with the cert

5. **Update Tailscale Ingress** to forward to kgateway port 443
   - Test if it does TCP passthrough to HTTPS backend

6. **Update DNS** - A record pointing `dev.argocd.synapse.mabbott.dev` to Tailscale IP

7. **Test** from phone and computer

## Open Questions

1. Why doesn't `tailscale.com/expose` annotation work from phone over DERP relay?
2. Can Tailscale Ingress do TCP passthrough to an HTTPS backend?
3. Do we need Cloudflare API token for DNS01 challenge?

## Current Working Setup

```
Phone → Tailscale → synapse-gateway-1.tail93695b.ts.net → Tailscale Ingress (TLS) → kgateway (HTTP) → HTTPRoute → ArgoCD
```

Only works with `.ts.net` hostname. Custom domain breaks TLS handshake.
