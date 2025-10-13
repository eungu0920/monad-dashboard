import { Reset } from "@radix-ui/themes";
import { Link } from "@tanstack/react-router";
import monadLogo from "../../assets/monad_logo.svg";
import { useMedia } from "react-use";
import styles from "./logo.module.css";
import { Flex, Text } from "@radix-ui/themes";

export default function Logo() {
  const isWideScreen = useMedia("(min-width: 1366px)");

  return (
    <Reset>
      <Link to="/">
        <Flex align="center" gap="2">
          <img
            className={styles.logo}
            src={monadLogo}
            alt="Monad"
            style={{ height: "26px" }}
          />
          {isWideScreen && (
            <Text size="5" weight="bold" style={{ color: "white" }}>
              Monad
            </Text>
          )}
        </Flex>
      </Link>
    </Reset>
  );
}
