import { useAppDispatch, useAppSelector } from "@/app/store";
import { initializeProjectDetailed } from "@/composites/ProjectDetailed";
import { SelectPeriod } from "@/features/SelectPeriod";
import { PerformanceModule } from "@/widgets/PerformanceModule";
import { PlatformMetrics } from "@/widgets/PlatformMetrics";
import { ProjectTitle } from "@/widgets/ProjectTitle";
import { ResourceLinks } from "@/widgets/ResourceLinks";
import { MetricList } from "@/widgets/MetricList";
import { FC, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useForceUser } from "@/entities/User";
import { SetSelectedUsers } from "@/features/SetSelectedUsers";
import { useMetricDataUpdate } from "@/entities/Metric/hooks/useMetricDataUpdate";
import { LoadingCircle } from "@/shared/ui/LoadingCircle";
import { GoToProjectSettings } from "@/features/Project";
import { shallowEqual } from "react-redux";

interface ProjectPageProps {}

const ProjectPage: FC<ProjectPageProps> = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const { project, resources, metrics } = useAppSelector(
    (state) => ({
      project: state.project.project,
      resources: state.resource.resources,
      metrics: state.metric.metrics,
    }),
    shallowEqual,
  );

  const isLoading = useAppSelector((state) => state.project.isLoading);

  useEffect(() => {
    if (id) {
      initializeProjectDetailed(dispatch, id);
    } else {
      navigate(import.meta.env.VITE_BASE_PATH);
    }
  }, []);

  useForceUser();

  useMetricDataUpdate();

  return (
    <div className="flex gap-9 mt-14">
      <div className="w-full">
        {isLoading && (
          <div className="h-[calc(100vh-100px)] w-full flex justify-center items-center">
            <LoadingCircle size={80} />
          </div>
        )}
        {!isLoading && project && (
          <>
            <ProjectTitle hint={project.description}>
              {project.name}
            </ProjectTitle>
            <div className="pt-8"></div>
            <ul className="flex flex-col gap-6">
              {resources.map((resource) => (
                <li key={resource.id}>
                  <PlatformMetrics resource={resource} key={resource.id}>
                    <SetSelectedUsers resourceId={resource.id} />
                    <div className="pt-8" />
                    <PerformanceModule resource={resource.id} />
                    <div className="pt-8" />
                    <MetricList
                      metrics={metrics.filter(
                        (metric) => metric.resource == resource.id,
                      )}
                    />
                  </PlatformMetrics>
                </li>
              ))}
            </ul>
          </>
        )}
      </div>
      <div className="min-w-[25%] flex flex-col">
        <GoToProjectSettings
          link={import.meta.env.VITE_BASE_URL + "project/" + id + "/settings"}
        />
        <div className="mt-16" />
        <SelectPeriod />
        <div className="pt-20" />
        <ResourceLinks />
      </div>
    </div>
  );
};

export default ProjectPage;
