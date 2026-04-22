class GlitchDev < Formula
  desc "gl1tch workflow engine (HEAD — latest from main)"
  homepage "https://8op.org"
  license "MIT"
  head "https://github.com/8op-org/gl1tch.git", branch: "main"

  depends_on "babashka"

  def install
    mkdir_p share/"glitch/src/glitch"
    cp_r Dir["bb/src/glitch/*"], share/"glitch/src/glitch/"
    (share/"glitch/providers").install Dir["bb/providers/*.clj"]
    (share/"glitch").install "bb/resources/ast-grep-rules"

    (bin/"glitch").write <<~SH
      #!/bin/bash
      exec bb -cp "#{share}/glitch/src" -m glitch.main "$@"
    SH
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
