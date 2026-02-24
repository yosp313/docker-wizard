package dockerfile

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
		startCommandBlock("/app/app")
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
		installCmd = "npm ci"
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
		startCommandBlock(startCmd)
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
		startCommandBlock("python main.py")
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
		startCommandBlock("ruby app.rb")
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
		startCommandBlock("php -S 0.0.0.0:8080 -t public")
}

func dockerfileJava(details LanguageDetails) string {
	javaVersion := versionOrDefault(details.JavaVersion, "21")

	buildStage := ""
	if details.HasPomXML {
		buildStage = "" +
			"FROM maven:3.9-eclipse-temurin-" + javaVersion + " AS build\n" +
			"WORKDIR /src\n" +
			"COPY pom.xml ./\n" +
			"RUN mvn -q -DskipTests dependency:go-offline || true\n" +
			"COPY . .\n" +
			"RUN mvn -q -DskipTests package && mkdir -p /out && cp \"$(find target -maxdepth 1 -type f -name '*.jar' | head -n 1)\" /out/app.jar\n"
	} else if details.HasGradle || details.HasGradleKts {
		buildStage = "" +
			"FROM gradle:8-jdk" + javaVersion + " AS build\n" +
			"WORKDIR /src\n" +
			"COPY . .\n" +
			"RUN if [ -f ./gradlew ]; then chmod +x ./gradlew && ./gradlew build -x test; else gradle build -x test; fi && mkdir -p /out && cp \"$(find build/libs -maxdepth 1 -type f -name '*.jar' | head -n 1)\" /out/app.jar\n"
	} else {
		buildStage = "" +
			"FROM eclipse-temurin:" + javaVersion + "-jre AS build\n" +
			"WORKDIR /src\n" +
			"COPY . .\n" +
			"RUN mkdir -p /out && if [ -f app.jar ]; then cp app.jar /out/app.jar; fi\n"
	}

	return "" +
		buildStage +
		"\n" +
		"FROM eclipse-temurin:" + javaVersion + "-jre\n" +
		"WORKDIR /app\n" +
		"COPY --from=build /out/app.jar /app/app.jar\n" +
		"EXPOSE 8080\n" +
		startCommandBlock("java -jar /app/app.jar")
}

func dockerfileDotNet(details LanguageDetails) string {
	dotnetVersion := versionOrDefault(details.DotNetVersion, "8.0")
	entryDll := "app.dll"
	if details.DotNetProject != "" {
		entryDll = details.DotNetProject + ".dll"
	}

	return "" +
		"FROM mcr.microsoft.com/dotnet/sdk:" + dotnetVersion + " AS build\n" +
		"WORKDIR /src\n" +
		"COPY *.csproj ./\n" +
		"RUN dotnet restore || true\n" +
		"COPY . .\n" +
		"RUN dotnet publish -c Release -o /out\n" +
		"\n" +
		"FROM mcr.microsoft.com/dotnet/aspnet:" + dotnetVersion + "\n" +
		"WORKDIR /app\n" +
		"COPY --from=build /out/ ./\n" +
		"EXPOSE 8080\n" +
		startCommandBlock("dotnet /app/"+entryDll)
}

func dockerfileGeneric() string {
	return "" +
		"FROM alpine:3.20\n" +
		"WORKDIR /app\n" +
		"COPY . .\n" +
		startCommandBlock("sh")
}

func versionOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func startCommandBlock(defaultCommand string) string {
	return "" +
		"ENV APP_START_CMD=\"" + defaultCommand + "\"\n" +
		"CMD [\"sh\", \"-lc\", \"$APP_START_CMD\"]\n"
}
