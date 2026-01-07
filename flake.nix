{
  description = "Cortex - Synapse Platform Control Plane";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";

    axon = {
      url = "path:../axon";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, axon }:
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
            echo ""
            echo "Commands:"
            echo "  just fleet-up       Create Kind clusters"
            echo "  just fleet-down     Destroy clusters"
            echo "  just fleet-status   Check health"
            echo "  just bootstrap      Bootstrap Flux"
            echo "  just validate       Validate manifests"
            echo ""
          '';
        };

        checks.default = pkgs.runCommand "validate-cortex" {
          buildInputs = [ pkgs.kubeconform pkgs.kustomize ];
        } ''
          cd ${self}
          for dir in deploy/*/; do
            if [ -f "$dir/kustomization.yaml" ]; then
              echo "Validating $dir..."
              kustomize build "$dir" | kubeconform -strict -skip CustomResourceDefinition
            fi
          done
          touch $out
        '';
      }
    );
}
