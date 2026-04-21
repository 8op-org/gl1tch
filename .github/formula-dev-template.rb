class GlitchDev < Formula
  desc "gl1tch workflow engine (HEAD — latest from main)"
  homepage "https://8op.org"
  license "MIT"
  head "https://github.com/8op-org/gl1tch.git", branch: "main"

  depends_on "babashka"

  def install
    cd "bb" do
      system "bb", "build"
      bin.install "build/glitch"
    end
    (share/"glitch/providers").install Dir["bb/providers/*.clj"]
  end

  def post_install
    provider_dir = etc/"glitch/providers"
    provider_dir.mkpath
    (share/"glitch/providers").each_child do |f|
      provider_dir.install_symlink f unless (provider_dir/f.basename).exist?
    end
  end

  test do
    assert_match "glitch", shell_output("#{bin}/glitch version")
  end
end
