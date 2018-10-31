package crawler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

const (
	// DefaultOutputFileDot is the .dot file location to save the sitemap graph information to.
	DefaultOutputFileDot = "sitemap.dot"

	// DefaultOutputFileSvg is the .svg file location to save the sitemap graph to.
	DefaultOutputFileSvg = "sitemap.svg"
)

// Text renders the given sitemap as a list of pages and links found.
func Text(s Sitemap) (string, error) {
	edges := getEdges(s)

	var buffer bytes.Buffer

	_, err := buffer.WriteString("\npages:\n\n")
	if err != nil {
		return "", fmt.Errorf("error generating the text output: %s", err.Error())
	}

	for url := range s {
		_, err := buffer.WriteString(string(url) + "\n")
		if err != nil {
			return "", fmt.Errorf("error writing the page list: %s", err.Error())
		}
	}

	_, err = buffer.WriteString("\nlinks:\n\n")
	if err != nil {
		return "", fmt.Errorf("error generating the text output: %s", err.Error())
	}

	for _, edge := range edges {
		_, err := buffer.WriteString(fmt.Sprintf("%s -> %s\n", edge[0], edge[1]))
		if err != nil {
			return "", fmt.Errorf("error writing the links: %s", err.Error())
		}
	}

	return buffer.String(), nil
}

// Graph renders the given sitemap as a graph saved in an SVG file.
// The graph is generated using dot, a graphviz tool.
// The dot command is invoked using the exec command, and it is assumed that dot is already installed.
// The sitemap data is first saved as a .dot file, which is then passed as source to the dot command.
func Graph(s Sitemap) error {
	f, err := os.Create(DefaultOutputFileDot)
	if err != nil {
		return fmt.Errorf("error creating the .dot output file writer: %s", err.Error())
	}

	defer f.Close()

	err = writeDot(f, s)
	if err != nil {
		return fmt.Errorf("error generating the dot file: %s", err.Error())
	}

	cmd := exec.Command("dot", "-Tsvg", DefaultOutputFileDot, "-o", DefaultOutputFileSvg)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error generating the svg file: %s", err.Error())
	}

	return nil
}

func writeDot(writer io.Writer, sitemap Sitemap) (err error) {

	edges := getEdges(sitemap)

	w := bufio.NewWriter(writer)

	_, err = w.WriteString("digraph G {\n")
	if err != nil {
		return err
	}

	for _, edge := range edges {
		_, err = w.WriteString(fmt.Sprintf(`"%s"->"%s";`, edge[0], edge[1]))
		if err != nil {
			return err
		}

		err := w.WriteByte('\n')
		if err != nil {
			return err
		}
	}

	for url := range sitemap {
		_, err := w.WriteString(fmt.Sprintf(`"%s";`, url))
		if err != nil {
			return err
		}

		err = w.WriteByte('\n')
		if err != nil {
			return err
		}
	}

	_, err = w.WriteString("}\n")
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil
}

func getEdges(sitemap Sitemap) [][2]string {
	var edges [][2]string

	for page, links := range sitemap {
		for _, link := range links {
			edges = append(edges, [2]string{string(page), link})
		}
	}

	return edges
}
