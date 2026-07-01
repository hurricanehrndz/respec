{ pkgs, ... }:

{
  # https://devenv.sh/packages/
  packages = [
    pkgs.git
    pkgs.hugo
    pkgs.just
    pkgs.golangci-lint
  ];

  # https://devenv.sh/languages/
  languages.go.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
