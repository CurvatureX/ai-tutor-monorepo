import type { ReactNode } from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";

import styles from "./index.module.css";

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx("hero hero--primary", styles.heroBanner)}>
      <div className="container">
        <h1 className="hero__title">{siteConfig.title}</h1>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/tech_design/架构设计"
          >
            View Tech Design
          </Link>
        </div>
      </div>
    </header>
  );
}

function ReadmeContent() {
  return (
    <section className={styles.readme}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <div className={styles.readmeContent}>
              <h1>AI Tutor Monorepo</h1>

              <p>
                A comprehensive AI-powered tutoring platform built with
                microservices architecture, supporting multiple client
                applications and real-time learning experiences.
              </p>

              <h2>🏗️ Architecture Overview</h2>

              <p>
                This monorepo contains a complete AI tutoring ecosystem with:
              </p>

              <ul>
                <li>
                  <strong>Multiple Client Applications</strong>: Flutter mobile
                  app and Unity 3D immersive learning environments
                </li>
                <li>
                  <strong>Microservices Backend</strong>: Scalable services for
                  user management, conversations, speech processing, analytics,
                  and notifications
                </li>
                <li>
                  <strong>Shared Infrastructure</strong>: Common protocols,
                  configurations, and utilities
                </li>
                <li>
                  <strong>DevOps & Deployment</strong>: Kubernetes, Helm charts,
                  monitoring, and CI/CD pipelines
                </li>
              </ul>

              <h2>📱 Client Applications</h2>

              <div className="row">
                <div className="col col--6">
                  <h3>Flutter Mobile App</h3>
                  <ul>
                    <li>
                      Cross-platform mobile application for iOS and Android
                    </li>
                    <li>Real-time chat interface with AI tutors</li>
                    <li>Voice interaction capabilities</li>
                    <li>Progress tracking and analytics</li>
                  </ul>
                </div>
                <div className="col col--6">
                  <h3>Unity 3D Application</h3>
                  <ul>
                    <li>Immersive 3D learning environments</li>
                    <li>Interactive educational content</li>
                    <li>Gamified learning experiences</li>
                    <li>VR/AR support capabilities</li>
                  </ul>
                </div>
              </div>

              <h2>🔧 Backend Services</h2>

              <div className="row">
                <div className="col col--6">
                  <h3>Core Services</h3>
                  <ul>
                    <li>
                      <strong>User Service</strong>: Authentication, user
                      profiles, and account management
                    </li>
                    <li>
                      <strong>Conversation Service</strong>: AI-powered chat and
                      tutoring logic
                    </li>
                    <li>
                      <strong>Speech Service</strong>: Voice recognition and
                      text-to-speech processing
                    </li>
                    <li>
                      <strong>Analytics Service</strong>: Learning progress
                      tracking and insights
                    </li>
                    <li>
                      <strong>Notification Service</strong>: Real-time alerts
                      and messaging
                    </li>
                  </ul>
                </div>
                <div className="col col--6">
                  <h3>Infrastructure</h3>
                  <ul>
                    <li>
                      <strong>API Gateway</strong>: Request routing and load
                      balancing
                    </li>
                    <li>
                      <strong>Shared Protocols</strong>: gRPC definitions and
                      common data models
                    </li>
                    <li>
                      <strong>Monitoring</strong>: Observability and performance
                      tracking
                    </li>
                    <li>
                      <strong>Database</strong>: User data, conversation
                      history, and analytics storage
                    </li>
                  </ul>
                </div>
              </div>

              <h2>🚀 Quick Start</h2>

              <h3>Prerequisites</h3>
              <ul>
                <li>Docker and Docker Compose</li>
                <li>Kubernetes cluster (for production)</li>
                <li>Go 1.19+ (for gateway and notification service)</li>
                <li>Python 3.9+ (for AI services)</li>
                <li>Flutter SDK (for mobile development)</li>
                <li>Unity 2022.3+ (for 3D application)</li>
              </ul>

              <h3>Development Setup</h3>
              <ol>
                <li>
                  <strong>Clone the repository</strong>
                  <pre>
                    <code>
                      git clone
                      https://github.com/your-org/ai-tutor-monorepo.git{"\n"}cd
                      ai-tutor-monorepo
                    </code>
                  </pre>
                </li>
                <li>
                  <strong>Start backend services</strong>
                  <pre>
                    <code>docker-compose up -d</code>
                  </pre>
                </li>
                <li>
                  <strong>Run Flutter app</strong>
                  <pre>
                    <code>
                      cd clients/flutter-app{"\n"}flutter pub get{"\n"}flutter
                      run
                    </code>
                  </pre>
                </li>
                <li>
                  <strong>Open Unity project</strong>
                  <pre>
                    <code># Open clients/unity-3d in Unity Editor</code>
                  </pre>
                </li>
              </ol>

              <h2>📚 Documentation</h2>

              <div className="row">
                <div className="col col--4">
                  <div className={styles.docCard}>
                    <h3>
                      <Link to="/tech_design/架构设计">Tech Design</Link>
                    </h3>
                    <p>Architecture decisions and system design</p>
                  </div>
                </div>
                <div className="col col--4">
                  <div className={styles.docCard}>
                    <h3>
                      <Link to="/api_documentation/intro">
                        API Documentation
                      </Link>
                    </h3>
                    <p>Service APIs and integration guides</p>
                  </div>
                </div>
                <div className="col col--4">
                  <div className={styles.docCard}>
                    <h3>Infrastructure</h3>
                    <p>Kubernetes and cloud deployment instructions</p>
                  </div>
                </div>
              </div>

              <h2>🛠️ Development</h2>

              <h3>Project Structure</h3>
              <pre>
                <code>{`├── clients/           # Client applications
│   ├── flutter-app/   # Mobile application
│   └── unity-3d/      # 3D learning environment
├── services/          # Backend microservices
├── gateway/           # API gateway
├── shared/            # Common code and protocols
├── infrastructure/    # Deployment configurations
└── tools/             # Development and build tools`}</code>
              </pre>

              <div className="margin-top--lg">
                <p>
                  <strong>CurvTech Inc.</strong> - Building the future of
                  AI-powered education
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

export default function Home(): ReactNode {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout>
      <HomepageHeader />
      <main>
        <ReadmeContent />
      </main>
    </Layout>
  );
}
