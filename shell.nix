{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go_1_23
    gopls
    gotools
    golangci-lint
  ];

  shellHook = ''
    echo "Denver development environment"
    echo "Go version: $(go version)"
    echo ""
    echo "Commands:"
    echo "  go build -o denver  # Build binary"
    echo "  ./denver --help     # Run binary"
    echo ""
  '';
}
