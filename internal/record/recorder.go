// Package record implements code to record all links and broken links from the checker
package record

// Recorder receives links and broken links from the checker.
// Its main purpose is to safely collect these from multiple go routines at the same time.
type Recorder struct {
	recordLink       chan Link
	recordBrokenLink chan BrokenLink
	stop             chan bool
	result           chan Recording
}

func NewRecorder() *Recorder {
	c := &Recorder{
		// A buffer of 10 should be enough space as recorder is a light routine.
		recordLink:       make(chan Link, 10),       //nolint:gomnd
		recordBrokenLink: make(chan BrokenLink, 10), //nolint:gomnd
		stop:             make(chan bool, 1),
		result:           make(chan Recording, 1),
	}

	go func() {
		// Init variables
		// Randomly chosen sizes to limit initial reallocation
		links := make([]Link, 0, 64)                          //nolint:gomnd
		brokenLinks := make(map[string]BrokenLinkDetails, 16) //nolint:gomnd

		// Start recording loop
		for {
			select {
			case link := <-c.recordLink:
				links = append(links, link)
			case brokenLink := <-c.recordBrokenLink:
				brokenLinks[brokenLink.AbsoluteURL] = brokenLink.BrokenLinkDetails
			case <-c.stop:
				// Ensure buffered channels are empty
				for {
					select {
					case link := <-c.recordLink:
						links = append(links, link)
					case brokenLink := <-c.recordBrokenLink:
						brokenLinks[brokenLink.AbsoluteURL] = brokenLink.BrokenLinkDetails
					default:
						// All channels are empty => Return result
						c.result <- Recording{
							Links:              links,
							BrokenLinkByAbsURL: brokenLinks,
						}
						return
					}
				}
			}
		}
	}()

	return c
}

func (c *Recorder) RecordLink(link Link) {
	c.recordLink <- link
}

func (c *Recorder) RecordBrokenLink(brokenLink BrokenLink) {
	c.recordBrokenLink <- brokenLink
}

func (c *Recorder) Stop() Recording {
	c.stop <- true
	return <-c.result
}
