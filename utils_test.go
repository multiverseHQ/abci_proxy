package abciproxy

import "sync"

//slow and super ugly, but really its a sprint
type InterceptedMethod struct {
	Name          string
	Calls         [][]interface{}
	wg            sync.WaitGroup
	wgLock        sync.Mutex
	expectedCalls int
	currentCalls  int
}

func NewInterceptedMethod(name string) InterceptedMethod {
	return InterceptedMethod{
		Name:   name,
		Calls:  nil,
		wg:     sync.WaitGroup{},
		wgLock: sync.Mutex{},
	}
}

func (im *InterceptedMethod) Notify(args ...interface{}) {
	im.Calls = append(im.Calls, append([]interface{}{}, args...))

	im.wgLock.Lock()
	defer im.wgLock.Unlock()
	if im.expectedCalls == 0 {
		return
	}
	im.currentCalls++
	if im.expectedCalls == im.currentCalls {
		im.wg.Done()
		im.expectedCalls = 0
	}

}

func (im *InterceptedMethod) ExpectCall(N int) {
	im.wgLock.Lock()
	defer im.wgLock.Unlock()

	im.expectedCalls = N
	im.currentCalls = 0
	im.wg.Add(1)
}

func (im *InterceptedMethod) WaitForExpected() {
	im.wg.Wait()
}
