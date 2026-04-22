class Glitch < Formula
  desc "gl1tch workflow engine"
  homepage "https://8op.org"
  license "MIT"
  version "__VERSION__"

  depends_on "babashka"

  on_macos do
    url "__BASE_URL__/glitch___VERSION___darwin_arm64.tar.gz"
    sha256 "__DARWIN_ARM64_SHA__"
  end

  on_linux do
    if Hardware::CPU.arm?
      url "__BASE_URL__/glitch___VERSION___linux_arm64.tar.gz"
      sha256 "__LINUX_ARM64_SHA__"
    else
      url "__BASE_URL__/glitch___VERSION___linux_amd64.tar.gz"
      sha256 "__LINUX_AMD64_SHA__"
    end
  end

  def install
    bin.install "glitch"
    (share/"glitch").install "ast-grep-rules"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/glitch version")
  end
end
