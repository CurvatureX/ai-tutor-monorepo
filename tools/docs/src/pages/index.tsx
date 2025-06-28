import type { ReactNode } from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import { useEffect, useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

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

  return (
    <section className={styles.readme}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <div className={styles.readmeContent}>
              <ReactMarkdown
                remarkPlugins={[remarkGfm]}
                components={{
                  // Custom components for better styling
                  h1: ({ children }) => (
                    <h1 className="markdown-h1">{children}</h1>
                  ),
                  h2: ({ children }) => (
                    <h2 className="markdown-h2">{children}</h2>
                  ),
                  h3: ({ children }) => (
                    <h3 className="markdown-h3">{children}</h3>
                  ),
                  code: (props) => {
                    const { children, className, ...rest } = props;
                    const match = /language-(\w+)/.exec(className || "");
                    return (
                      <code className="markdown-inline-code" {...rest}>
                        {children}
                      </code>
                    );
                  },
                  pre: ({ children }) => (
                    <pre className="markdown-pre">{children}</pre>
                  ),
                  ul: ({ children }) => (
                    <ul className="markdown-ul">{children}</ul>
                  ),
                  ol: ({ children }) => (
                    <ol className="markdown-ol">{children}</ol>
                  ),
                  li: ({ children }) => (
                    <li className="markdown-li">{children}</li>
                  ),
                  p: ({ children }) => <p className="markdown-p">{children}</p>,
                  strong: ({ children }) => (
                    <strong className="markdown-strong">{children}</strong>
                  ),
                  em: ({ children }) => (
                    <em className="markdown-em">{children}</em>
                  ),
                }}
              >
                {readmeContent}
              </ReactMarkdown>
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
