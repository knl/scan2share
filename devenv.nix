{ pkgs, ... }:

{
  # https://devenv.sh/basics/
  # env.GREET = "devenv";

  # https://devenv.sh/packages/
  packages = [ pkgs.gopls pkgs.go-tools ];

  # https://devenv.sh/scripts/
  # scripts.hello.exec = "echo hello from $GREET";

  # enterShell = ''
  # '';

  # https://devenv.sh/languages/
  languages.go.enable = true;

  # https://devenv.sh/pre-commit-hooks/
  # pre-commit.hooks.shellcheck.enable = true;

  # https://devenv.sh/processes/
  # processes.ping.exec = "ping example.com";
}
