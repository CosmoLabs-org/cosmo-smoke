package detector

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func touchNew(t *testing.T, dir, name string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("touch %s: %v", name, err)
	}
	f.Close()
}

func TestDetect_Java(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "pom.xml")
	types := Detect(dir)
	if !slices.Contains(types, Java) {
		t.Errorf("java project: expected Java in %v", types)
	}
}

func TestDetect_JavaGradle(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "package.json")
	touchNew(t, dir, "build.gradle")
	types := Detect(dir)
	if !slices.Contains(types, JavaGradle) {
		t.Errorf("java-gradle project: expected JavaGradle in %v", types)
	}
}

func TestDetect_DotNet(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "App.csproj")
	types := Detect(dir)
	if !slices.Contains(types, DotNet) {
		t.Errorf("dotnet project: expected DotNet in %v", types)
	}
}

func TestDetect_DotNet_Sln(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "MyApp.sln")
	types := Detect(dir)
	if !slices.Contains(types, DotNet) {
		t.Errorf("dotnet sln project: expected DotNet in %v", types)
	}
}

func TestDetect_Ruby(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "Gemfile")
	types := Detect(dir)
	if !slices.Contains(types, Ruby) {
		t.Errorf("ruby project: expected Ruby in %v", types)
	}
}

func TestDetect_PHP(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "composer.json")
	types := Detect(dir)
	if !slices.Contains(types, PHP) {
		t.Errorf("php project: expected PHP in %v", types)
	}
}

func TestDetect_Deno(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "deno.json")
	types := Detect(dir)
	if !slices.Contains(types, Deno) {
		t.Errorf("deno project: expected Deno in %v", types)
	}
}

func TestDetect_DenoJsonc(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "deno.jsonc")
	types := Detect(dir)
	if !slices.Contains(types, Deno) {
		t.Errorf("deno jsonc project: expected Deno in %v", types)
	}
}

func TestDetect_Terraform(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "main.tf")
	types := Detect(dir)
	if !slices.Contains(types, Terraform) {
		t.Errorf("terraform project: expected Terraform in %v", types)
	}
}

func TestDetect_Helm(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "Chart.yaml")
	types := Detect(dir)
	if !slices.Contains(types, Helm) {
		t.Errorf("helm project: expected Helm in %v", types)
	}
}

func TestDetect_Kustomize(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "kustomization.yaml")
	types := Detect(dir)
	if !slices.Contains(types, Kustomize) {
		t.Errorf("kustomize project: expected Kustomize in %v", types)
	}
}

func TestDetect_Serverless(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "serverless.yml")
	types := Detect(dir)
	if !slices.Contains(types, Serverless) {
		t.Errorf("serverless project: expected Serverless in %v", types)
	}
}

func TestDetect_Zig(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "build.zig")
	types := Detect(dir)
	if !slices.Contains(types, Zig) {
		t.Errorf("zig project: expected Zig in %v", types)
	}
}

func TestDetect_Elixir(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "mix.exs")
	types := Detect(dir)
	if !slices.Contains(types, Elixir) {
		t.Errorf("elixir project: expected Elixir in %v", types)
	}
}

func TestDetect_Scala(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "build.sbt")
	types := Detect(dir)
	if !slices.Contains(types, Scala) {
		t.Errorf("scala project: expected Scala in %v", types)
	}
}

func TestDetect_SwiftServer(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "Package.swift")
	types := Detect(dir)
	if !slices.Contains(types, SwiftServer) {
		t.Errorf("swift-server project: expected SwiftServer in %v", types)
	}
}

func TestDetect_SwiftServer_IosExcluded(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "Package.swift")
	touchNew(t, dir, "MyApp.xcodeproj") // Should trigger iOS, not SwiftServer
	os.MkdirAll(filepath.Join(dir, "MyApp.xcodeproj"), 0o755)
	types := Detect(dir)
	if slices.Contains(types, SwiftServer) {
		t.Errorf("swift-server should not be detected when xcodeproj exists, got %v", types)
	}
}

func TestDetect_DartServer(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte("name: myapp\ndependencies:\n  shelf: ^1.0.0\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	types := Detect(dir)
	if !slices.Contains(types, DartServer) {
		t.Errorf("dart-server project: expected DartServer in %v", types)
	}
}

func TestDetect_DartServer_FlutterExcluded(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte("name: myapp\ndependencies:\n  flutter:\n    sdk: flutter\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	types := Detect(dir)
	if slices.Contains(types, DartServer) {
		t.Errorf("dart-server should not be detected when flutter dep exists, got %v", types)
	}
}

func TestDetect_Hugo(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "hugo.toml")
	types := Detect(dir)
	if !slices.Contains(types, Hugo) {
		t.Errorf("hugo project: expected Hugo in %v", types)
	}
}

func TestDetect_Hugo_ConfigToml(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "config.toml")
	os.MkdirAll(filepath.Join(dir, "content"), 0o755)
	types := Detect(dir)
	if !slices.Contains(types, Hugo) {
		t.Errorf("hugo config.toml project: expected Hugo in %v", types)
	}
}

func TestDetect_Astro(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "astro.config.mjs")
	types := Detect(dir)
	if !slices.Contains(types, Astro) {
		t.Errorf("astro project: expected Astro in %v", types)
	}
}

func TestDetect_Jekyll(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "_config.yml")
	touchNew(t, dir, "Gemfile")
	types := Detect(dir)
	if !slices.Contains(types, Jekyll) {
		t.Errorf("jekyll project: expected Jekyll in %v", types)
	}
}

func TestDetect_Jekyll_NoGemfile(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "_config.yml")
	types := Detect(dir)
	if slices.Contains(types, Jekyll) {
		t.Errorf("jekyll should not be detected without Gemfile, got %v", types)
	}
}

func TestDetect_Make(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "Makefile")
	types := Detect(dir)
	if !slices.Contains(types, Make) {
		t.Errorf("make project: expected Make in %v", types)
	}
}

func TestDetect_CMake(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "CMakeLists.txt")
	types := Detect(dir)
	if !slices.Contains(types, CMake) {
		t.Errorf("cmake project: expected CMake in %v", types)
	}
}

func TestDetect_Haskell(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "stack.yaml")
	types := Detect(dir)
	if !slices.Contains(types, Haskell) {
		t.Errorf("haskell project: expected Haskell in %v", types)
	}
}

func TestDetect_Haskell_Cabal(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "myproject.cabal")
	types := Detect(dir)
	if !slices.Contains(types, Haskell) {
		t.Errorf("haskell cabal project: expected Haskell in %v", types)
	}
}

func TestDetect_Lua(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "myproject-1.0-1.rockspec")
	types := Detect(dir)
	if !slices.Contains(types, Lua) {
		t.Errorf("lua project: expected Lua in %v", types)
	}
}

func TestGenerateConfig_Java(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{Java})
	if len(cfg.Tests) < 2 {
		t.Errorf("java config: expected at least 2 tests, got %d", len(cfg.Tests))
	}
	if len(cfg.Prereqs) < 1 {
		t.Errorf("java config: expected at least 1 prereq, got %d", len(cfg.Prereqs))
	}
}

func TestGenerateConfig_DotNet(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{DotNet})
	if len(cfg.Tests) < 2 {
		t.Errorf("dotnet config: expected at least 2 tests, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Terraform(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{Terraform})
	if len(cfg.Tests) < 2 {
		t.Errorf("terraform config: expected at least 2 tests, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Helm(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{Helm})
	if len(cfg.Tests) < 2 {
		t.Errorf("helm config: expected at least 2 tests, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Elixir(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{Elixir})
	if len(cfg.Tests) < 3 {
		t.Errorf("elixir config: expected at least 3 tests, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Zig(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{Zig})
	if len(cfg.Tests) < 2 {
		t.Errorf("zig config: expected at least 2 tests, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Make(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{Make})
	if len(cfg.Tests) < 1 {
		t.Errorf("make config: expected at least 1 test, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_CMake(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{CMake})
	if len(cfg.Tests) < 2 {
		t.Errorf("cmake config: expected at least 2 tests, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_PHP(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{PHP})
	if len(cfg.Tests) < 2 {
		t.Errorf("php config: expected at least 2 tests, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Deno(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{Deno})
	if len(cfg.Tests) < 2 {
		t.Errorf("deno config: expected at least 2 tests, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Ruby(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig(dir, []ProjectType{Ruby})
	if len(cfg.Tests) < 1 {
		t.Errorf("ruby config: expected at least 1 test, got %d", len(cfg.Tests))
	}
}

func TestGenerateConfig_Ruby_WithRakefile(t *testing.T) {
	dir := t.TempDir()
	touchNew(t, dir, "Rakefile")
	cfg := GenerateConfig(dir, []ProjectType{Ruby})
	if len(cfg.Tests) < 2 {
		t.Errorf("ruby config with Rakefile: expected at least 2 tests, got %d", len(cfg.Tests))
	}
}
