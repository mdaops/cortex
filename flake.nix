{
  description = "Cortex - Synapse Platform Control Plane";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Kubernetes
            kubectl
            kubernetes-helm
            kustomize
            kind

            # GitOps
            fluxcd

            # Development
            just
            jq
            yq-go
            kubeconform

            # Git
            git
            gh

            # Utilities
            k9s
            kubectx
            stern
          ];

          shellHook = ''
            echo "Cortex - Synapse Platform Control Plane"
          '';
        };
      }
    );
}
