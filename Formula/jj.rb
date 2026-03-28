class Jj < Formula
  desc "jump jump — directory bookmarking tool"
  homepage "https://github.com/Patrik-Stas/jj"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Patrik-Stas/jj/releases/download/v#{version}/_jj-darwin-arm64"
      sha256 "PLACEHOLDER"

      def install
        bin.install "_jj-darwin-arm64" => "_jj"
      end
    else
      url "https://github.com/Patrik-Stas/jj/releases/download/v#{version}/_jj-darwin-amd64"
      sha256 "PLACEHOLDER"

      def install
        bin.install "_jj-darwin-amd64" => "_jj"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Patrik-Stas/jj/releases/download/v#{version}/_jj-linux-arm64"
      sha256 "PLACEHOLDER"

      def install
        bin.install "_jj-linux-arm64" => "_jj"
      end
    else
      url "https://github.com/Patrik-Stas/jj/releases/download/v#{version}/_jj-linux-amd64"
      sha256 "PLACEHOLDER"

      def install
        bin.install "_jj-linux-amd64" => "_jj"
      end
    end
  end

  def caveats
    <<~EOS
      Add the following to your shell config:

        For zsh (~/.zshrc):
          eval "$(_jj init zsh)"

        For bash (~/.bashrc):
          eval "$(_jj init bash)"
    EOS
  end

  test do
    assert_match "jj — jump jump", shell_output("#{bin}/_jj init zsh")
  end
end
