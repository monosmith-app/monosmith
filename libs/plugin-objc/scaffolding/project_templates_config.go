package scaffolding

// templates are every project's: the podspec, named for its pod as CocoaPods
// expects, and one class with its header. The homepage, source and author are
// placeholders on a domain that does not resolve; the author replaces them.
var templates = map[string]string{
	"[[.Dir]]/[[.Module]].podspec": `# smithKind: [[.Kind]]
[[- if .Stack]]
# smithStack: [[.Stack]]
[[- end]]
Pod::Spec.new do |s|
  s.name         = '[[.Module]]'
  s.version      = '0.0.0'
  s.summary      = 'A [[.Kind]] of the [[.Workspace]] workspace.'
  s.homepage     = 'https://example.invalid/[[.Module]]'
  s.license      = { :type => 'MIT' }
  s.author       = { 'Author' => 'author@example.invalid' }
  s.source       = { :git => 'https://example.invalid/[[.Module]].git', :tag => s.version.to_s }
  s.ios.deployment_target = '12.0'
  s.osx.deployment_target = '10.13'
  s.source_files = '[[.Module]]/Classes/**/*.{h,m}'
end
`,
	"[[.Dir]]/[[.Module]]/Classes/[[.Module]].h": `#import <Foundation/Foundation.h>

NS_ASSUME_NONNULL_BEGIN

/// Greets people.
@interface [[.Module]] : NSObject

/// Returns a greeting for the given name.
+ (NSString *)greet:(NSString *)who;

@end

NS_ASSUME_NONNULL_END
`,
	"[[.Dir]]/[[.Module]]/Classes/[[.Module]].m": `#import "[[.Module]].h"

@implementation [[.Module]]

+ (NSString *)greet:(NSString *)who {
    return [NSString stringWithFormat:@"Hello, %@!", who];
}

@end
`,
}
