import type { ReactNode } from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import { useEffect, useState } from "react";

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
  const [readmeContent, setReadmeContent] = useState<string>("");
  const { siteConfig } = useDocusaurusContext();

  useEffect(() => {
    // Fetch the README.md from the static folder
    fetch(`${siteConfig.baseUrl}README.md`)
      .then((response) => response.text())
      .then((text) => setReadmeContent(text))
      .catch(() => {
        setReadmeContent("# README.md content could not be loaded");
      });
  }, [siteConfig.baseUrl]);

  // Simple markdown to HTML converter for basic elements
  const markdownToHtml = (markdown: string) => {
    return markdown
      .replace(/^# (.*$)/gim, "<h1>$1</h1>")
      .replace(/^## (.*$)/gim, "<h2>$1</h2>")
      .replace(/^### (.*$)/gim, "<h3>$1</h3>")
      .replace(/^\* (.*$)/gim, "<li>$1</li>")
      .replace(/^- (.*$)/gim, "<li>$1</li>")
      .replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>")
      .replace(/\*(.*?)\*/g, "<em>$1</em>")
      .replace(/```([\s\S]*?)```/g, "<pre><code>$1</code></pre>")
      .replace(/`(.*?)`/g, "<code>$1</code>")
      .replace(/^\d+\. (.*$)/gim, "<li>$1</li>")
      .replace(/\n\n/g, "</p><p>")
      .replace(/^(?!<[h|l|p])/gm, "<p>")
      .replace(/(?!<\/[h|l|p])$/gm, "</p>")
      .replace(/<p><\/p>/g, "")
      .replace(/<p>(<[h|l])/g, "$1")
      .replace(/(<\/[h|l]>)<\/p>/g, "$1")
      .replace(/<li>/g, "<ul><li>")
      .replace(/<\/li>/g, "</li></ul>")
      .replace(/<\/ul><ul>/g, "");
  };

  return (
    <section className={styles.readme}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <div
              className={styles.readmeContent}
              dangerouslySetInnerHTML={{
                __html: markdownToHtml(readmeContent),
              }}
            />
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
