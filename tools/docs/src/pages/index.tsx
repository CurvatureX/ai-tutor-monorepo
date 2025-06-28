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
          <Link className="button button--primary button--lg" to="/readme">
            View Full README
          </Link>
        </div>
      </div>
    </header>
  );
}

function ProjectOverview() {
  return (
    <section className={styles.readme}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <div className={styles.readmeContent}>
              <h2>🏗️ Architecture Overview</h2>
              <p>
                A comprehensive AI-powered tutoring platform built with
                microservices architecture, supporting multiple client
                applications and real-time learning experiences.
              </p>

              <div className="row">
                <div className="col col--6">
                  <h3>📱 Client Applications</h3>
                  <ul>
                    <li>
                      <strong>Flutter Mobile App</strong> - Cross-platform
                      mobile application
                    </li>
                    <li>
                      <strong>Unity 3D Application</strong> - Immersive learning
                      environments
                    </li>
                  </ul>
                </div>
                <div className="col col--6">
                  <h3>🔧 Backend Services</h3>
                  <ul>
                    <li>
                      <strong>User Service</strong> - Authentication and
                      profiles
                    </li>
                    <li>
                      <strong>Conversation Service</strong> - AI-powered chat
                      logic
                    </li>
                    <li>
                      <strong>Speech Service</strong> - Voice processing
                    </li>
                    <li>
                      <strong>Analytics Service</strong> - Learning insights
                    </li>
                  </ul>
                </div>
              </div>

              <div className="text--center" style={{ marginTop: "2rem" }}>
                <Link
                  className="button button--outline button--primary button--lg"
                  to="/readme"
                >
                  Read Complete Documentation →
                </Link>
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
        <ProjectOverview />
      </main>
    </Layout>
  );
}
