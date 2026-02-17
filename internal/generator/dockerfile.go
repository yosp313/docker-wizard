package generator

import "fmt"

func Dockerfile(details LanguageDetails) (string, error) {
	switch details.Type {
	case LanguageGo:
		return dockerfileGo(details), nil
	case LanguageNode:
		return dockerfileNode(details), nil
	case LanguagePython:
		return dockerfilePython(details), nil
	case LanguageRuby:
		return dockerfileRuby(details), nil
	case LanguagePHP:
		return dockerfilePHP(details), nil
	case LanguageJava:
		return dockerfileJava(details), nil
	case LanguageDotNet:
		return dockerfileDotNet(details), nil
	case LanguageUnknown:
		return dockerfileGeneric(), nil
	default:
		return "", fmt.Errorf("unsupported language: %s", details.Type)
	}
}

func dockerfileGo(details LanguageDetails) string {
	moduleFiles := "COPY go.mod ./\n"
	if details.HasGoSum {
		moduleFiles = "COPY go.mod go.sum ./\n"
	}
	goVersion := versionOrDefault(details.GoVersion, "1.25")

	return "" +
		"FROM golang:" + goVersion + "-alpine AS build\n" +
		"WORKDIR /src\n" +
		moduleFiles +
		"RUN go mod download\n" +
		"COPY . .\n" +
		"RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app .\n" +
		"\n" +
		"FROM alpine:3.20\n" +
		"WORKDIR /app\n" +
		"COPY --from=build /out/app /app/app\n" +
		"EXPOSE 8080\n" +
		"CMD [\"/app/app\"]\n"
}

func dockerfileNode(details LanguageDetails) string {
	installCmd := "npm install"
	startCmd := "npm start"
	lockCopy := ""
	corepack := ""
	nodeVersion := versionOrDefault(details.NodeVersion, "20")

	if details.HasYarnLock {
		installCmd = "yarn install --frozen-lockfile"
		startCmd = "yarn start"
		lockCopy = "COPY yarn.lock ./\n"
		corepack = "RUN corepack enable\n"
	} else if details.HasPnpmLock {
		installCmd = "pnpm install --frozen-lockfile"
		startCmd = "pnpm start"
		lockCopy = "COPY pnpm-lock.yaml ./\n"
		corepack = "RUN corepack enable\n"
	} else if details.HasPackageLock {
		lockCopy = "COPY package-lock.json ./\n"
	}

	return "" +
		"FROM node:" + nodeVersion + "-alpine\n" +
		"WORKDIR /app\n" +
		"COPY package.json ./\n" +
		lockCopy +
		corepack +
		"RUN " + installCmd + "\n" +
		"COPY . .\n" +
		"EXPOSE 8080\n" +
		"CMD [\"" + splitCmd(startCmd)[0] + "\"" + joinCmdArgs(startCmd) + "]\n"
}

func dockerfilePython(details LanguageDetails) string {
	install := ""
	copy := "COPY . .\n"
	if details.HasRequirements {
		install = "COPY requirements.txt ./\nRUN pip install --no-cache-dir -r requirements.txt\n"
	}
	pythonVersion := versionOrDefault(details.PythonVersion, "3.12")

	return "" +
		"FROM python:" + pythonVersion + "-slim\n" +
		"WORKDIR /app\n" +
		install +
		copy +
		"EXPOSE 8080\n" +
		"CMD [\"python\", \"main.py\"]\n"
}

func dockerfileRuby(details LanguageDetails) string {
	install := ""
	if details.HasGemfile {
		install = "COPY Gemfile ./\n"
		if details.HasGemfileLock {
			install += "COPY Gemfile.lock ./\n"
		}
		install += "RUN bundle install\n"
	}
	rubyVersion := versionOrDefault(details.RubyVersion, "3.3")

	return "" +
		"FROM ruby:" + rubyVersion + "-alpine\n" +
		"WORKDIR /app\n" +
		install +
		"COPY . .\n" +
		"EXPOSE 8080\n" +
		"CMD [\"ruby\", \"app.rb\"]\n"
}

func dockerfilePHP(details LanguageDetails) string {
	phpVersion := versionOrDefault(details.PHPVersion, "8.3")
	composerCopy := ""
	if details.HasComposerJSON {
		composerCopy = "COPY composer.json ./\n"
	}

	return "" +
		"FROM php:" + phpVersion + "-fpm-alpine\n" +
		"WORKDIR /app\n" +
		composerCopy +
		"COPY . .\n" +
		"EXPOSE 8080\n" +
		"CMD [\"php\", \"-S\", \"0.0.0.0:8080\", \"-t\", \"public\"]\n"
}

func dockerfileJava(details LanguageDetails) string {
	javaVersion := versionOrDefault(details.JavaVersion, "21")
	return "" +
		"FROM eclipse-temurin:" + javaVersion + "-jre\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		"EXPOSE 8080\n" +
		"CMD [\"java\", \"-jar\", \"app.jar\"]\n"
}

func dockerfileDotNet(details LanguageDetails) string {
	dotnetVersion := versionOrDefault(details.DotNetVersion, "8.0")
	return "" +
		"FROM mcr.microsoft.com/dotnet/aspnet:" + dotnetVersion + "\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		"EXPOSE 8080\n" +
		"CMD [\"dotnet\", \"app.dll\"]\n"
}

func dockerfileGeneric() string {
	return "" +
		"FROM alpine:3.20\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		"CMD [\"sh\"]\n"
}

func splitCmd(cmd string) []string {
	return splitWords(cmd)
}

func versionOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func joinCmdArgs(cmd string) string {
	parts := splitWords(cmd)
	if len(parts) <= 1 {
		return ""
	}

	out := ""
	for _, part := range parts[1:] {
		out += ", \"" + part + "\""
	}
	return out
}

func splitWords(value string) []string {
	var parts []string
	current := ""
	for _, r := range value {
		if r == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
			continue
		}
		current += string(r)
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
